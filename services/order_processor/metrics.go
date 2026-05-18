package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	OrdersEnqueuedTotal prometheus.Counter
	OrdersDequeuedTotal prometheus.Counter

	ProcessingSpawnedTotal prometheus.Counter

	QueueDepth             prometheus.Gauge
	OldestQueuedAgeSeconds prometheus.Gauge
	InFlight               prometheus.Gauge
	PaymentCallsInFlight   prometheus.Gauge

	OrdersProcessedTotal *prometheus.CounterVec

	QueueWaitDuration   prometheus.Histogram
	ProcessingDuration  prometheus.Histogram
	PaymentCallDuration prometheus.Histogram

	LastProcessedTimestamp prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		OrdersEnqueuedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "order_processor_orders_enqueued_total",
			Help: "Total orders admitted into the processor queue.",
		}),
		OrdersDequeuedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "order_processor_orders_dequeued_total",
			Help: "Total orders removed from the processor queue for processing.",
		}),
		ProcessingSpawnedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "order_processor_processing_spawned_total",
			Help: "Total processing goroutines spawned after dequeue.",
		}),

		QueueDepth: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "order_processor_queue_depth",
			Help: "Current number of orders waiting in the processor queue.",
		}),
		OldestQueuedAgeSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "order_processor_oldest_queued_age_seconds",
			Help: "Age in seconds of the oldest order still waiting in the processor queue.",
		}),
		InFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "order_processor_in_flight",
			Help: "Number of orders currently being processed.",
		}),
		PaymentCallsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "order_processor_payment_calls_in_flight",
			Help: "Concurrent payment service calls currently in progress.",
		}),

		OrdersProcessedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_processor_orders_processed_total",
				Help: "Orders completed by outcome.",
			},
			[]string{"outcome"},
		),

		QueueWaitDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "order_processor_queue_wait_duration_seconds",
			Help:    "Time an order spent waiting in queue before processing began.",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
		}),
		ProcessingDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "order_processor_processing_duration_seconds",
			Help:    "End-to-end time spent processing an order after dequeue.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60},
		}),
		PaymentCallDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "order_processor_payment_call_duration_seconds",
			Help:    "Time spent calling the payment service.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60},
		}),

		LastProcessedTimestamp: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "order_processor_last_processed_timestamp_seconds",
			Help: "Unix timestamp of the last successfully completed order.",
		}),
	}

	reg.MustRegister(
		m.OrdersEnqueuedTotal,
		m.OrdersDequeuedTotal,
		m.ProcessingSpawnedTotal,

		m.QueueDepth,
		m.OldestQueuedAgeSeconds,
		m.InFlight,
		m.PaymentCallsInFlight,

		m.OrdersProcessedTotal,

		m.QueueWaitDuration,
		m.ProcessingDuration,
		m.PaymentCallDuration,

		m.LastProcessedTimestamp,
	)

	return m
}
