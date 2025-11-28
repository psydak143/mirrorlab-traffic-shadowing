package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg := LoadConfig()
	initMetrics()

	guard := NewGuardrail(GuardrailConfig{
		LatencyRatioThreshold: cfg.LatencyRatioThreshold,
		ErrorDeltaThreshold:   cfg.ErrorDeltaThreshold,
		BreachConsecutive:     cfg.BreachConsecutive,
		Cooldown:              cfg.Cooldown,
	})

	proxy, err := NewProxy(cfg, guard)
	if err != nil {
		log.Fatalf("failed to create proxy: %v", err)
	}

	mux := http.NewServeMux()

	// Health & metrics
	mux.HandleFunc("/healthz", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Control endpoints
	mux.Handle("/control/status", controlStatusHandler(guard))
	mux.Handle("/control/enable", controlEnableHandler(guard))
	mux.Handle("/control/disable", controlDisableHandler(guard))

	// Everything else goes through the proxy
	mux.Handle("/", proxy)

	log.Printf("Starting proxy on %s (baseline=%s candidate=%s mirrorFraction=%.2f)",
		cfg.ListenAddr, cfg.BaselineURL, cfg.CandidateURL, cfg.MirrorFraction)

	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		log.Fatal(err)
	}
}
