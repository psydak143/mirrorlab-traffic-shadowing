package main

import (
	"fmt"
	"sync"
	"time"
)

type GuardrailConfig struct {
	LatencyRatioThreshold float64
	ErrorDeltaThreshold   float64
	BreachConsecutive     int
	Cooldown              time.Duration
}

type GuardrailStatus struct {
	Enabled         bool      `json:"enabled"`
	LastAbortTime   time.Time `json:"lastAbortTime"`
	LastAbortReason string    `json:"lastAbortReason"`
	BreachStreak    int       `json:"breachStreak"`

	LatencyRatioThreshold float64       `json:"latencyRatioThreshold"`
	ErrorDeltaThreshold   float64       `json:"errorDeltaThreshold"`
	BreachConsecutive     int           `json:"breachConsecutive"`
	Cooldown              time.Duration `json:"cooldownSeconds"`
}

type Guardrail struct {
	mu    sync.Mutex
	cfg   GuardrailConfig
	state struct {
		enabled         bool
		lastAbortTime   time.Time
		lastAbortReason string
		breachStreak    int
	}
}

func NewGuardrail(cfg GuardrailConfig) *Guardrail {
	g := &Guardrail{cfg: cfg}
	g.state.enabled = true
	mirrorEnabledGauge.Set(1)
	return g
}

// ShouldMirror returns whether mirroring is currently allowed (respecting cooldown).
func (g *Guardrail) ShouldMirror() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.state.enabled {
		return true
	}

	if g.cfg.Cooldown <= 0 {
		return false
	}

	if time.Since(g.state.lastAbortTime) >= g.cfg.Cooldown {
		// cooldown expired – re-enable
		g.state.enabled = true
		mirrorEnabledGauge.Set(1)
		return true
	}

	return false
}

// ObservePair looks at one baseline/candidate pair and may trigger an abort
// if the candidate is consistently worse.
func (g *Guardrail) ObservePair(route string, baselineLat, candidateLat time.Duration, baselineErr, candidateErr bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.state.enabled {
		return
	}

	// Avoid division by zero; treat tiny baseline as 1ms.
	if baselineLat <= 0 {
		baselineLat = time.Millisecond
	}

	latencyRatio := float64(candidateLat) / float64(baselineLat)
	breach := false

	if latencyRatio > g.cfg.LatencyRatioThreshold {
		breach = true
	}

	// Very simple "error delta": candidate has error but baseline doesn't.
	if candidateErr && !baselineErr {
		// error delta is effectively 1 - 0 = 1; if threshold < 1, it's a breach.
		if 1.0 > g.cfg.ErrorDeltaThreshold {
			breach = true
		}
	}

	if breach {
		g.state.breachStreak++
	} else if g.state.breachStreak > 0 {
		g.state.breachStreak--
	}

	if g.state.breachStreak >= g.cfg.BreachConsecutive {
		// Abort
		g.state.enabled = false
		g.state.lastAbortTime = time.Now()
		g.state.lastAbortReason = fmt.Sprintf(
			"route=%s latencyRatio=%.2f baselineErr=%v candidateErr=%v",
			route, latencyRatio, baselineErr, candidateErr,
		)
		g.state.breachStreak = 0

		mirrorEnabledGauge.Set(0)
		mirrorAbortsTotal.Inc()
	}
}

func (g *Guardrail) Disable(reason string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.state.enabled {
		return
	}
	g.state.enabled = false
	g.state.lastAbortTime = time.Now()
	if reason == "" {
		reason = "manual disable"
	}
	g.state.lastAbortReason = reason
	g.state.breachStreak = 0

	mirrorEnabledGauge.Set(0)
	mirrorAbortsTotal.Inc()
}

func (g *Guardrail) Enable(reason string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.state.enabled {
		return
	}
	g.state.enabled = true
	g.state.breachStreak = 0
	if reason != "" {
		g.state.lastAbortReason = reason
	}
	mirrorEnabledGauge.Set(1)
}

func (g *Guardrail) Status() GuardrailStatus {
	g.mu.Lock()
	defer g.mu.Unlock()

	return GuardrailStatus{
		Enabled:         g.state.enabled,
		LastAbortTime:   g.state.lastAbortTime,
		LastAbortReason: g.state.lastAbortReason,
		BreachStreak:    g.state.breachStreak,

		LatencyRatioThreshold: g.cfg.LatencyRatioThreshold,
		ErrorDeltaThreshold:   g.cfg.ErrorDeltaThreshold,
		BreachConsecutive:     g.cfg.BreachConsecutive,
		Cooldown:              g.cfg.Cooldown,
	}
}
