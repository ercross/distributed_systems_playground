package chaos

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func Inject(cfg *Config, metrics *Metrics, operation string) error {
	if cfg == nil {
		return nil
	}

	inc := func(kind string) {
		if metrics != nil {
			metrics.InjectedTotal.WithLabelValues(kind, operation).Inc()
		}
	}

	if cfg.CrashEnabled && rand.Float64() < cfg.CrashRate {
		inc("crash")
		log.Printf("[chaos/%s] crash injected during operation=%s", cfg.ServiceName, operation)
		os.Exit(1)
	}

	if cfg.LatencyEnabled {
		delta := cfg.MaxLatencyMs - cfg.MinLatencyMs
		if delta < 0 {
			delta = 0
		}
		sleep := cfg.MinLatencyMs + rand.Intn(delta+1)
		inc("latency")
		log.Printf("[chaos/%s] latency injected: %dms during operation=%s", cfg.ServiceName, sleep, operation)
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}

	if cfg.TimeoutEnabled && rand.Float64() < cfg.TimeoutRate {
		inc("timeout")
		log.Printf("[chaos/%s] timeout injected: sleeping %dms during operation=%s", cfg.ServiceName, cfg.TimeoutMs, operation)
		time.Sleep(time.Duration(cfg.TimeoutMs) * time.Millisecond)
		return fmt.Errorf("chaos timeout injected in operation=%s after %dms", operation, cfg.TimeoutMs)
	}

	if cfg.FailureEnabled && rand.Float64() < cfg.FailureRate {
		inc("failure")
		reason := randomFailureReason()
		log.Printf("[chaos/%s] failure injected during operation=%s: %s", cfg.ServiceName, operation, reason)
		return fmt.Errorf("chaos failure injected in operation=%s: %s", operation, reason)
	}

	return nil
}
