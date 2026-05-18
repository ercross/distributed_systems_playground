package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"shared/chaos"
	"time"

	"shared/models"
)

type queuedOrder struct {
	Order      models.Order
	EnqueuedAt time.Time
}

type Processor struct {
	paymentURL   string
	httpClient   *http.Client
	queue        *UnboundedQueue[queuedOrder]
	metrics      *Metrics
	chaosMetrics *chaos.Metrics
	chaosConfig  *chaos.Config
}

func NewProcessor(
	chaosCfg *chaos.Config,
	chaosMetrics *chaos.Metrics,
	paymentURL string,
	metrics *Metrics,
) *Processor {
	p := &Processor{
		chaosConfig:  chaosCfg,
		chaosMetrics: chaosMetrics,
		paymentURL:   paymentURL,
		metrics:      metrics,
		httpClient: &http.Client{
			Timeout: 0,
		},
		queue: NewUnboundedQueue[queuedOrder](),
	}

	go p.refreshQueueAgeMetric()

	return p
}

// Start begins the dispatcher loop.
// Deliberately unsafe:
// - dequeue forever
// - spawn a goroutine per order
// - no worker pool
func (p *Processor) Start() {
	go func() {
		for {
			item := p.queue.DequeueBlocking()

			p.metrics.QueueDepth.Dec()
			p.metrics.OrdersDequeuedTotal.Inc()
			p.metrics.ProcessingSpawnedTotal.Inc()

			go p.processOrder(&item.Order, item.EnqueuedAt)
		}
	}()
}

func (p *Processor) Enqueue(order models.Order) {
	item := queuedOrder{
		Order:      order,
		EnqueuedAt: time.Now(),
	}

	p.queue.Enqueue(item)

	p.metrics.QueueDepth.Inc()
	p.metrics.OrdersEnqueuedTotal.Inc()
}

func (p *Processor) QueueDepth() int {
	return p.queue.Len()
}

func (p *Processor) processOrder(order *models.Order, enqueuedAt time.Time) {
	p.metrics.InFlight.Inc()
	defer p.metrics.InFlight.Dec()

	p.metrics.QueueWaitDuration.Observe(time.Since(enqueuedAt).Seconds())

	start := time.Now()
	defer func() {
		p.metrics.ProcessingDuration.Observe(time.Since(start).Seconds())
	}()

	log.Printf("[order-processor] processing order=%s amount=%.2f", order.ID, order.Amount)

	// Internal/post-admission chaos:
	// the request was already accepted and queued,
	// but work can still fail, stall, or crash here.
	if err := chaos.Inject(p.chaosConfig, p.chaosMetrics, "process_order"); err != nil {
		log.Printf("[order-processor] chaos injected for order=%s: %v", order.ID, err)
		p.metrics.OrdersProcessedTotal.WithLabelValues("chaos_error").Inc()
		return
	}

	// Simulate some business logic work
	time.Sleep(10 * time.Millisecond)

	payStart := time.Now()
	payResp, err := p.callPaymentService(order)
	p.metrics.PaymentCallDuration.Observe(time.Since(payStart).Seconds())

	if err != nil {
		log.Printf("[order-processor] payment call error order=%s: %v", order.ID, err)
		p.metrics.OrdersProcessedTotal.WithLabelValues("payment_failed").Inc()
		return
	}

	if payResp.Status != "success" {
		log.Printf(
			"[order-processor] payment declined order=%s tx=%s: %s",
			order.ID,
			payResp.TransactionID,
			payResp.Error,
		)
		p.metrics.OrdersProcessedTotal.WithLabelValues("payment_failed").Inc()
		return
	}

	p.metrics.OrdersProcessedTotal.WithLabelValues("success").Inc()
	p.metrics.LastProcessedTimestamp.SetToCurrentTime()

	log.Printf(
		"[order-processor] order complete order=%s tx=%s duration=%.2fs",
		order.ID,
		payResp.TransactionID,
		time.Since(start).Seconds(),
	)
}

func (p *Processor) callPaymentService(order *models.Order) (*models.PaymentResponse, error) {
	p.metrics.PaymentCallsInFlight.Inc()
	defer p.metrics.PaymentCallsInFlight.Dec()

	req := models.PaymentRequest{
		OrderID: order.ID,
		UserID:  order.UserID,
		Amount:  order.Amount,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		p.paymentURL+"/pay",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payResp models.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payResp); err != nil {
		return nil, err
	}

	return &payResp, nil
}

func (p *Processor) refreshQueueAgeMetric() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if p.queue.Len() == 0 {
			p.metrics.OldestQueuedAgeSeconds.Set(0)
			continue
		}

		item, ok := p.queue.Peek()
		if !ok {
			p.metrics.OldestQueuedAgeSeconds.Set(0)
			continue
		}

		p.metrics.OldestQueuedAgeSeconds.Set(time.Since(item.EnqueuedAt).Seconds())
	}
}
