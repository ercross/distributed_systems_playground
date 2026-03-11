package main

import (
	"encoding/json"
	"net/http"

	"shared/models"
	"shared/utils"
)

type Server struct {
	processor *Processor
}

func NewServer(processor *Processor) *Server {
	return &Server{processor: processor}
}

func (s *Server) handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
		return
	}

	s.processor.Enqueue(order)

	utils.WriteJSON(w, http.StatusAccepted, map[string]string{
		"status":   "queued",
		"order_id": order.ID,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"status":      "ok",
		"service":     "order-processor",
		"queue_depth": s.processor.QueueDepth(),
	})
}
