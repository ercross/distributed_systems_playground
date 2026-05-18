package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	ServiceUp prometheus.Gauge

	PaymentRequestsTotal prometheus.Counter
	PaymentOutcomesTotal *prometheus.CounterVec

	PaymentProcessingDuration prometheus.Histogram
	BankGatewayCallDuration   prometheus.Histogram

	TransactionsInFlight     prometheus.Gauge
	BankGatewayCallsInFlight prometheus.Gauge
	BankHangsInFlight        prometheus.Gauge
	BankHangEnabled          prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		ServiceUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "payment_service_up",
			Help: "Whether the payment service process is running and able to report liveness.",
		}),
		PaymentRequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "payment_service_requests_total",
			Help: "Total payment requests admitted into the payment handler.",
		}),
		PaymentOutcomesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_service_outcomes_total",
				Help: "Payment outcomes by type.",
			},
			[]string{"outcome"},
		),
		PaymentProcessingDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "payment_service_processing_duration_seconds",
			Help:    "End-to-end time spent inside payment processing for completed requests.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60},
		}),
		BankGatewayCallDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "payment_service_bank_gateway_duration_seconds",
			Help:    "Time spent in the simulated bank gateway call for completed or timed-out calls.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60},
		}),
		TransactionsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "payment_service_transactions_in_flight",
			Help: "Number of payment requests currently being processed.",
		}),
		BankGatewayCallsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "payment_service_bank_gateway_calls_in_flight",
			Help: "Number of requests currently inside simulated bank gateway processing.",
		}),
		BankHangsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "payment_service_bank_hangs_in_flight",
			Help: "Number of payment requests currently stuck in the simulated unbounded bank hang.",
		}),
		BankHangEnabled: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "payment_service_bank_hang_enabled",
			Help: "Whether simulated unbounded bank hangs are enabled for this process.",
		}),
	}

	reg.MustRegister(
		m.ServiceUp,
		m.PaymentRequestsTotal,
		m.PaymentOutcomesTotal,
		m.PaymentProcessingDuration,
		m.BankGatewayCallDuration,
		m.TransactionsInFlight,
		m.BankGatewayCallsInFlight,
		m.BankHangsInFlight,
		m.BankHangEnabled,
	)

	return m
}
