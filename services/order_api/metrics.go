package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	OrdersAcceptedTotal prometheus.Counter
	OrdersRejectedTotal prometheus.Counter

	ProcessorForwardInFlight prometheus.Gauge
	ProcessorForwardDuration prometheus.Histogram
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		OrdersAcceptedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "order_api_orders_accepted_total",
			Help: "Total orders accepted by the edge for asynchronous forwarding.",
		}),
		OrdersRejectedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "order_api_orders_rejected_total",
			Help: "Total order creation requests rejected by the edge before acceptance.",
		}),
		ProcessorForwardInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "order_api_processor_forward_in_flight",
			Help: "Number of concurrent goroutines currently trying to forward accepted orders to the processor.",
		}),
		ProcessorForwardDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "order_api_processor_forward_duration_seconds",
			Help:    "Time spent attempting to forward an accepted order to the processor.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60},
		}),
	}

	reg.MustRegister(
		m.OrdersAcceptedTotal,
		m.OrdersRejectedTotal,
		m.ProcessorForwardInFlight,
		m.ProcessorForwardDuration,
	)

	return m
}
