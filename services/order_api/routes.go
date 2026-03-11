package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"shared/chaos"
	"shared/httpx"
)

func (s *Server) routes(red *httpx.REDMetrics, chaosCfg *chaos.Config, chaosMetrics *chaos.Metrics) http.Handler {
	r := chi.NewRouter()

	// Apply RED metrics to all application routes.
	r.Use(httpx.Middleware(red))

	// /orders gets chaos middleware because we want to simulate
	// request-path pain specifically for order creation.
	r.With(
		chaos.Middleware(*chaosCfg, chaosMetrics, "create_order"),
	).Post("/orders", s.handleCreateOrder)

	return r
}
