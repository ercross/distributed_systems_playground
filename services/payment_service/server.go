package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"shared/chaos"
	"shared/models"
	"shared/utils"
)

type Server struct {
	chaos        *chaos.Config
	chaosMetrics *chaos.Metrics
	payment      PaymentConfig
	metrics      *Metrics
	reqID        uint64
}

func NewServer(
	chaosCfg *chaos.Config,
	chaosMetrics *chaos.Metrics,
	paymentCfg PaymentConfig,
	metrics *Metrics,
) *Server {
	metrics.ServiceUp.Set(1)

	if paymentCfg.BankHangEnabled {
		metrics.BankHangEnabled.Set(1)
	} else {
		metrics.BankHangEnabled.Set(0)
	}

	return &Server{
		chaos:        chaosCfg,
		chaosMetrics: chaosMetrics,
		payment:      paymentCfg,
		metrics:      metrics,
	}
}

func (s *Server) handlePay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	defer func() {
		s.metrics.PaymentProcessingDuration.Observe(time.Since(start).Seconds())
	}()

	s.metrics.TransactionsInFlight.Inc()
	defer s.metrics.TransactionsInFlight.Dec()

	s.metrics.PaymentRequestsTotal.Inc()

	id := atomic.AddUint64(&s.reqID, 1)

	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.metrics.PaymentOutcomesTotal.WithLabelValues("invalid_request").Inc()
		utils.WriteJSON(w, http.StatusBadRequest, models.PaymentResponse{
			OrderID: req.OrderID,
			Status:  "error",
			Error:   "invalid request body",
		})
		return
	}

	log.Printf(
		"[payment-service] processing req=%d order=%s user=%s amount=%.2f",
		id,
		req.OrderID,
		req.UserID,
		req.Amount,
	)

	// Post-admission/internal chaos:
	// the request already reached the handler; now simulate internal service pain.
	if err := chaos.Inject(s.chaos, s.chaosMetrics, "process_payment"); err != nil {
		s.metrics.PaymentOutcomesTotal.WithLabelValues("chaos_error").Inc()
		log.Printf("[payment-service] chaos error req=%d order=%s: %v", id, req.OrderID, err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.PaymentResponse{
			OrderID: req.OrderID,
			Status:  "error",
			Error:   err.Error(),
		})
		return
	}

	s.metrics.BankGatewayCallsInFlight.Inc()
	defer s.metrics.BankGatewayCallsInFlight.Dec()

	// 0) Optional unbounded hang mode
	if s.payment.BankHangEnabled && rand.Float64() < s.payment.BankHangRate {
		s.metrics.BankHangsInFlight.Inc()
		defer s.metrics.BankHangsInFlight.Dec()

		s.metrics.PaymentOutcomesTotal.WithLabelValues("bank_hang").Inc()
		log.Printf("[payment-service] bank HANG injected req=%d order=%s", id, req.OrderID)

		select {}
	}

	// 1) Simulated bank gateway timeout
	bankStart := time.Now()
	if rand.Float64() < s.payment.BankTimeoutRate {
		log.Printf(
			"[payment-service] bank timeout injected req=%d order=%s sleeping=%dms",
			id,
			req.OrderID,
			s.payment.BankTimeoutMs,
		)

		time.Sleep(time.Duration(s.payment.BankTimeoutMs) * time.Millisecond)
		s.metrics.BankGatewayCallDuration.Observe(time.Since(bankStart).Seconds())
		s.metrics.PaymentOutcomesTotal.WithLabelValues("bank_timeout").Inc()

		utils.WriteJSON(w, http.StatusGatewayTimeout, models.PaymentResponse{
			OrderID: req.OrderID,
			Status:  "error",
			Error:   "bank gateway timeout",
		})
		return
	}

	// Simulate normal bank latency
	time.Sleep(time.Duration(20+rand.Intn(180)) * time.Millisecond)
	s.metrics.BankGatewayCallDuration.Observe(time.Since(bankStart).Seconds())

	// 2) Insufficient funds
	if rand.Float64() < s.payment.InsufficientFundsRate {
		log.Printf(
			"[payment-service] insufficient funds req=%d order=%s amount=%.2f",
			id,
			req.OrderID,
			req.Amount,
		)
		s.metrics.PaymentOutcomesTotal.WithLabelValues("insufficient_funds").Inc()

		utils.WriteJSON(w, http.StatusPaymentRequired, models.PaymentResponse{
			OrderID: req.OrderID,
			Status:  "declined",
			Error:   "insufficient funds",
		})
		return
	}

	// 3) Generic decline
	if rand.Float64() < s.payment.DeclineRate {
		reasons := []string{
			"card expired",
			"fraud risk score too high",
			"card reported lost",
			"do not honor",
			"card velocity exceeded",
		}
		reason := reasons[rand.Intn(len(reasons))]

		log.Printf("[payment-service] card declined req=%d order=%s: %s", id, req.OrderID, reason)
		s.metrics.PaymentOutcomesTotal.WithLabelValues("declined").Inc()

		utils.WriteJSON(w, http.StatusPaymentRequired, models.PaymentResponse{
			OrderID: req.OrderID,
			Status:  "declined",
			Error:   reason,
		})
		return
	}

	txID := uuid.New().String()
	s.metrics.PaymentOutcomesTotal.WithLabelValues("success").Inc()

	log.Printf(
		"[payment-service] payment success req=%d order=%s tx=%s amount=%.2f duration=%.2fs",
		id,
		req.OrderID,
		txID,
		req.Amount,
		time.Since(start).Seconds(),
	)

	utils.WriteJSON(w, http.StatusOK, models.PaymentResponse{
		OrderID:       req.OrderID,
		TransactionID: txID,
		Status:        "success",
		ProcessedAt:   time.Now(),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "payment-service",
		"config": map[string]any{
			"decline_rate":            s.payment.DeclineRate,
			"insufficient_funds_rate": s.payment.InsufficientFundsRate,
			"bank_timeout_rate":       s.payment.BankTimeoutRate,
			"bank_timeout_ms":         s.payment.BankTimeoutMs,
			"bank_hang_enabled":       s.payment.BankHangEnabled,
			"bank_hang_rate":          s.payment.BankHangRate,
		},
	})
}
