package main

import "github.com/prometheus/client_golang/prometheus"

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mirror_requests_total",
			Help: "Total requests handled by proxy, per route and target (baseline/candidate).",
		},
		[]string{"route", "target"},
	)

	latencyHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mirror_latency_seconds",
			Help:    "Latency of upstream calls in seconds, per route and target.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route", "target"},
	)

	diffMismatchesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mirror_diff_mismatches_total",
			Help: "Total number of response mismatches between baseline and candidate per route.",
		},
		[]string{"route"},
	)

	mirrorAbortsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mirror_aborts_total",
			Help: "Number of times mirroring was auto-disabled by guardrail.",
		},
	)

	mirrorEnabledGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mirror_enabled",
			Help: "1 if mirroring to candidate is enabled, 0 if disabled.",
		},
	)
)

func initMetrics() {
	prometheus.MustRegister(
		requestsTotal,
		latencyHistogram,
		diffMismatchesTotal,
		mirrorAbortsTotal,
		mirrorEnabledGauge,
	)
}
