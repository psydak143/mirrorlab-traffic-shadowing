package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	BaselineURL  string
	CandidateURL string

	MirrorFraction float64

	LatencyRatioThreshold float64
	ErrorDeltaThreshold   float64
	BreachConsecutive     int
	Cooldown              time.Duration

	ListenAddr string
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getenvFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		log.Printf("invalid %s=%q, using default %.3f", key, v, def)
		return def
	}
	return f
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("invalid %s=%q, using default %d", key, v, def)
		return def
	}
	return i
}

func getenvDurationSeconds(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("invalid %s=%q, using default %s", key, v, def)
		return def
	}
	return time.Duration(i) * time.Second
}

func LoadConfig() Config {
	return Config{
		BaselineURL:  getenv("BASELINE_URL", "http://service-v1:4001"),
		CandidateURL: getenv("CANDIDATE_URL", "http://service-v2:4002"),

		MirrorFraction: getenvFloat("MIRROR_FRACTION", 0.25),

		LatencyRatioThreshold: getenvFloat("ABORT_P99_RATIO", 1.15),    // >1.15x slower
		ErrorDeltaThreshold:   getenvFloat("ABORT_ERROR_DELTA", 0.005), // unused in this simplified guard, but kept for config
		BreachConsecutive:     getenvInt("ABORT_CONSECUTIVE", 5),
		Cooldown:              getenvDurationSeconds("ABORT_COOLDOWN", 10*time.Second),

		ListenAddr: getenv("LISTEN_ADDR", ":8080"),
	}
}
