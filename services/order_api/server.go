package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"shared/models"
	"shared/retry"
	"shared/utils"
	"time"
)

type Server struct {
	processorURL string
	httpClient   *http.Client
	metrics      *Metrics
	admission    *simpleAdmissionControl
}

func NewServer(processorURL string, metrics *Metrics, admission *simpleAdmissionControl) *Server {
	return &Server{
		processorURL: processorURL,
		metrics:      metrics,
		admission:    admission,
		httpClient: &http.Client{
			Timeout: 0,
		},
	}
}

func (s *Server) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.metrics.OrdersRejectedTotal.Inc()
		s.metrics.ValidationRejectionsTotal.Inc()
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON: " + err.Error(),
		})
		return
	}

	if req.UserID == "" || len(req.Items) == 0 {
		s.metrics.OrdersRejectedTotal.Inc()
		s.metrics.ValidationRejectionsTotal.Inc()
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "user_id and items are required",
		})
		return
	}

	// Admission check: must succeed before any order is created or work is spawned.
	if !s.admission.tryAcquireSlot() {
		s.metrics.OrdersRejectedTotal.Inc()
		s.metrics.AdmissionRejectionsTotal.Inc()
		s.metrics.AdmissionInUse.Set(float64(s.admission.slotsInUse()))
		utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "order admission temporarily unavailable",
		})
		return
	}
	s.metrics.AdmissionInUse.Set(float64(s.admission.slotsInUse()))

	var total float64
	for _, item := range req.Items {
		total += item.Price * float64(item.Quantity)
	}

	order := models.Order{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Amount:    total,
		Items:     req.Items,
		Status:    models.OrderStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.metrics.OrdersAcceptedTotal.Inc()

	log.Printf(
		"[order-api] order accepted id=%s user=%s amount=%.2f",
		order.ID,
		order.UserID,
		order.Amount,
	)

	go s.forwardToOrderProcessor(order)

	utils.WriteJSON(w, http.StatusCreated, order)
}

func (s *Server) forwardToOrderProcessor(order models.Order) {
	// Slot is held for the full forwarding lifecycle, including all retries and hangs.
	defer func() {
		s.admission.releaseSlot()
		s.metrics.AdmissionInUse.Set(float64(s.admission.slotsInUse()))
	}()

	s.metrics.ProcessorForwardInFlight.Inc()
	defer s.metrics.ProcessorForwardInFlight.Dec()

	start := time.Now()
	defer func() {
		s.metrics.ProcessorForwardDuration.Observe(time.Since(start).Seconds())
	}()

	body, err := json.Marshal(order)
	if err != nil {
		log.Printf("[order-api] failed to marshal order=%s: %v", order.ID, err)
		return
	}

	var resp *http.Response
	retryCfg := retry.ChaosConfig()

	err = retry.Do(context.Background(), retryCfg, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			s.processorURL+"/process",
			bytes.NewReader(body),
		)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = s.httpClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode >= 500 {
			defer resp.Body.Close()
			return fmt.Errorf("processor returned retryable 5xx: %d", resp.StatusCode)
		}

		return nil
	})
	if err != nil {
		log.Printf("[order-api] processor forward failed order=%s err=%v", order.ID, err)
		return
	}

	if resp != nil {
		defer resp.Body.Close()

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			log.Printf("[order-api] processor returned %d for order=%s", resp.StatusCode, order.ID)
			return
		}
	}

	log.Printf("[order-api] order forwarded to processor id=%s", order.ID)
}
