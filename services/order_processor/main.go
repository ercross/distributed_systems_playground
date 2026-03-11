package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"

	"shared/chaos"
	"shared/httpx"
	"shared/utils"
)

func main() {
	httpPort := utils.EnvOrDefault("HTTP_PORT", "8082")
	metricsPort := utils.EnvOrDefault("METRICS_PORT", "9092")
	paymentURL := utils.EnvOrDefault("PAYMENT_SERVICE_URL", "http://localhost:8083")

	reg := prometheus.DefaultRegisterer

	processorMetrics := NewMetrics(reg)
	redMetrics := httpx.NewREDMetrics(reg, "order_processor")
	chaosMetrics := chaos.NewMetrics(reg, "order_processor")
	chaosCfg := chaos.NewConfig("ORDER_PROCESSOR")

	proc := NewProcessor(chaosCfg, chaosMetrics, paymentURL, processorMetrics)
	proc.Start()

	srv := NewServer(proc)

	go func() {
		metricsRouter := chi.NewRouter()

		metricsRouter.Handle("/metrics", promhttp.Handler())
		metricsRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		log.Printf("[order-processor] metrics server on :%s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, metricsRouter); err != nil {
			log.Fatal(err)
		}
	}()

	appRouter := chi.NewRouter()

	// Apply RED metrics to all app routes.
	appRouter.Use(httpx.Middleware(redMetrics))

	// Only /process gets request-path chaos.
	appRouter.With(
		chaos.Middleware(*chaosCfg, chaosMetrics, "process_order"),
	).Post("/process", srv.handleProcess)

	appRouter.Get("/health", srv.handleHealth)

	log.Printf("[order-processor] HTTP server on :%s | payment=%s", httpPort, paymentURL)
	if err := http.ListenAndServe(":"+httpPort, appRouter); err != nil {
		log.Fatal(err)
	}
}
