package chaos

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func Middleware(cfg Config, metrics *Metrics, operation string) func(http.Handler) http.Handler {
	inc := func(kind string) {
		if metrics != nil {
			metrics.InjectedTotal.WithLabelValues(kind, operation).Inc()
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.CrashEnabled && rand.Float64() < cfg.CrashRate {
				inc("crash")
				log.Printf("[chaos/%s] crash injected on operation=%s", cfg.ServiceName, operation)
				os.Exit(1)
			}

			if cfg.PacketLossEnabled && rand.Float64() < cfg.PacketLossRate {
				inc("packet_loss")
				log.Printf("[chaos/%s] packet loss injected on operation=%s", cfg.ServiceName, operation)
				if hj, ok := w.(http.Hijacker); ok {
					conn, _, err := hj.Hijack()
					if err == nil {
						_ = conn.Close()
						return
					}
				}
				http.Error(w, "connection dropped", http.StatusServiceUnavailable)
				return
			}

			if cfg.LatencyEnabled {
				delta := cfg.MaxLatencyMs - cfg.MinLatencyMs
				if delta < 0 {
					delta = 0
				}
				sleep := cfg.MinLatencyMs + rand.Intn(delta+1)
				inc("latency")
				log.Printf("[chaos/%s] latency injected: %dms on operation=%s", cfg.ServiceName, sleep, operation)
				time.Sleep(time.Duration(sleep) * time.Millisecond)
			}

			if cfg.TimeoutEnabled && rand.Float64() < cfg.TimeoutRate {
				inc("timeout")
				log.Printf("[chaos/%s] timeout injected: sleeping %dms on operation=%s", cfg.ServiceName, cfg.TimeoutMs, operation)
				time.Sleep(time.Duration(cfg.TimeoutMs) * time.Millisecond)
			}

			if cfg.FailureEnabled && rand.Float64() < cfg.FailureRate {
				inc("failure")
				reason := randomFailureReason()
				log.Printf("[chaos/%s] failure injected: %s on operation=%s", cfg.ServiceName, reason, operation)
				http.Error(w, reason, http.StatusServiceUnavailable)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
