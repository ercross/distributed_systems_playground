package chaos

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	InjectedTotal *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer, service string) *Metrics {
	m := &Metrics{
		InjectedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: service + "_chaos_injected_total",
				Help: "Number of chaos events injected.",
			},
			[]string{"type", "operation"},
		),
	}
	reg.MustRegister(m.InjectedTotal)
	return m
}
