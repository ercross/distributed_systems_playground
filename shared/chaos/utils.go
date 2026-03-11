package chaos

import (
	"math/rand"
	"os"
	"strconv"
)

var failureReasons = []string{
	"connection refused by downstream",
	"database deadlock detected",
	"upstream returned 503",
	"disk I/O error",
	"out of memory",
	"network partition detected",
	"DNS resolution failed",
	"TLS handshake timeout",
}

func randomFailureReason() string {
	return failureReasons[rand.Intn(len(failureReasons))]
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func envFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}
