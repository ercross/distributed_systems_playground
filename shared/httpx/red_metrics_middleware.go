package httpx

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type REDMetrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	InFlight        prometheus.Gauge
}

func NewREDMetrics(reg prometheus.Registerer, service string) *REDMetrics {
	m := &REDMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: service + "_http_requests_total",
				Help: "Total HTTP requests handled.",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    service + "_http_request_duration_seconds",
				Help:    "HTTP request latency distribution.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		InFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: service + "_http_requests_in_flight",
				Help: "HTTP requests currently being handled.",
			},
		),
	}

	reg.MustRegister(
		m.RequestsTotal,
		m.RequestDuration,
		m.InFlight,
	)

	return m
}

func Middleware(metrics *REDMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			metrics.InFlight.Inc()
			defer metrics.InFlight.Dec()

			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)

			status := strconv.Itoa(rec.status)
			metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
			metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).
				Observe(time.Since(start).Seconds())
		})
	}
}
