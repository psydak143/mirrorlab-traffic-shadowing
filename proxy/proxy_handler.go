package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Proxy struct {
	baseline  *url.URL
	candidate *url.URL
	client    *http.Client
	guard     *Guardrail

	mirrorFraction float64
}

func NewProxy(cfg Config, guard *Guardrail) (*Proxy, error) {
	bu, err := url.Parse(cfg.BaselineURL)
	if err != nil {
		return nil, err
	}
	cu, err := url.Parse(cfg.CandidateURL)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	return &Proxy{
		baseline:  bu,
		candidate: cu,
		client:    client,
		guard:     guard,

		mirrorFraction: cfg.MirrorFraction,
	}, nil
}

func (p *Proxy) routeLabel(r *http.Request) string {
	// Simple and low-cardinality for this demo.
	return r.Method + " " + r.URL.Path
}

func (p *Proxy) targetURL(base *url.URL, src *url.URL) *url.URL {
	ref := &url.URL{
		Path:     src.Path,
		RawQuery: src.RawQuery,
	}
	return base.ResolveReference(ref)
}

func copyHeaders(dst, src http.Header) {
	for k, vals := range src {
		if strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
}

// ServeHTTP implements http.Handler for the proxy.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := p.routeLabel(r)

	// Read body once so we can reuse for both upstreams.
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	// ----- Baseline request (blocking; user response) -----
	bURL := p.targetURL(p.baseline, r.URL)
	bReq, err := http.NewRequestWithContext(r.Context(), r.Method, bURL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "failed to build baseline request", http.StatusInternalServerError)
		return
	}
	copyHeaders(bReq.Header, r.Header)

	start := time.Now()
	bResp, err := p.client.Do(bReq)
	bLatency := time.Since(start)

	requestsTotal.WithLabelValues(route, "baseline").Inc()

	if err != nil {
		log.Printf("baseline error %s %s: %v", r.Method, bURL.String(), err)
		latencyHistogram.WithLabelValues(route, "baseline").Observe(bLatency.Seconds())
		http.Error(w, "baseline upstream error", http.StatusBadGateway)
		return
	}
	defer bResp.Body.Close()

	bBody, err := io.ReadAll(bResp.Body)
	if err != nil {
		log.Printf("failed to read baseline body: %v", err)
		latencyHistogram.WithLabelValues(route, "baseline").Observe(bLatency.Seconds())
		http.Error(w, "failed to read baseline response", http.StatusBadGateway)
		return
	}
	latencyHistogram.WithLabelValues(route, "baseline").Observe(bLatency.Seconds())
	//baselineErr := bResp.StatusCode >= 500

	// Copy baseline headers/status/body to client.
	for k, vals := range bResp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(bResp.StatusCode)
	_, _ = w.Write(bBody)

	// ----- Candidate request (async mirror) -----
	if !p.guard.ShouldMirror() {
		return
	}
	if p.mirrorFraction <= 0 || rand.Float64() >= p.mirrorFraction {
		return
	}

	go p.mirrorToCandidate(route, r, bodyBytes, bBody, bResp.StatusCode, bLatency)
}

func (p *Proxy) mirrorToCandidate(route string, orig *http.Request, bodyBytes, baselineBody []byte, baselineStatus int, baselineLatency time.Duration) {
	cURL := p.targetURL(p.candidate, orig.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cReq, err := http.NewRequestWithContext(ctx, orig.Method, cURL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("candidate: failed to build request: %v", err)
		return
	}
	copyHeaders(cReq.Header, orig.Header)

	start := time.Now()
	cResp, err := p.client.Do(cReq)
	cLatency := time.Since(start)

	requestsTotal.WithLabelValues(route, "candidate").Inc()

	var candidateErr bool
	var cBody []byte

	if err != nil {
		log.Printf("candidate error %s %s: %v", orig.Method, cURL.String(), err)
		latencyHistogram.WithLabelValues(route, "candidate").Observe(cLatency.Seconds())
		candidateErr = true
	} else {
		defer cResp.Body.Close()
		cBody, err = io.ReadAll(cResp.Body)
		if err != nil {
			log.Printf("candidate: failed to read body: %v", err)
			candidateErr = true
		} else {
			candidateErr = cResp.StatusCode >= 500
		}
		latencyHistogram.WithLabelValues(route, "candidate").Observe(cLatency.Seconds())
	}

	// Diff baseline vs candidate bodies.
	if len(baselineBody) > 0 && len(cBody) > 0 {
		equal, err := equalJSONBodies(baselineBody, cBody)
		if err != nil {
			log.Printf("diff error route=%s: %v", route, err)
		} else if !equal {
			diffMismatchesTotal.WithLabelValues(route).Inc()
		}
	}

	p.guard.ObservePair(route, baselineLatency, cLatency, baselineStatus >= 500, candidateErr)
}

// Control handlers

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func controlStatusHandler(g *Guardrail) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, g.Status())
	}
}

func controlEnableHandler(g *Guardrail) http.HandlerFunc {
	type resp struct {
		Message string `json:"message"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		g.Enable("manual enable")
		writeJSON(w, http.StatusOK, resp{Message: "mirroring enabled"})
	}
}

func controlDisableHandler(g *Guardrail) http.HandlerFunc {
	type resp struct {
		Message string `json:"message"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		g.Disable("manual disable")
		writeJSON(w, http.StatusOK, resp{Message: "mirroring disabled"})
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
