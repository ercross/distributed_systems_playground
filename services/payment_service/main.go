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
	httpPort := utils.EnvOrDefault("HTTP_PORT", "8083")
	metricsPort := utils.EnvOrDefault("METRICS_PORT", "9093")

	reg := prometheus.DefaultRegisterer

	serviceMetrics := NewMetrics(reg)
	redMetrics := httpx.NewREDMetrics(reg, "payment_service")
	chaosMetrics := chaos.NewMetrics(reg, "payment_service")

	chaosCfg := chaos.NewConfig("PAYMENT_SERVICE")
	paymentCfg := loadPaymentConfig()

	if paymentCfg.BankHangEnabled {
		serviceMetrics.BankHangEnabled.Set(1)
	} else {
		serviceMetrics.BankHangEnabled.Set(0)
	}

	serviceMetrics.ServiceUp.Set(1)
	defer serviceMetrics.ServiceUp.Set(0)

	log.Printf(
		"[payment-service] payment config: decline=%.0f%% insufficient=%.0f%% bank_timeout=%.0f%%(%dms) bank_hang=%v(%.2f%%)",
		paymentCfg.DeclineRate*100,
		paymentCfg.InsufficientFundsRate*100,
		paymentCfg.BankTimeoutRate*100,
		paymentCfg.BankTimeoutMs,
		paymentCfg.BankHangEnabled,
		paymentCfg.BankHangRate*100,
	)

	srv := NewServer(chaosCfg, chaosMetrics, paymentCfg, serviceMetrics)

	go func() {
		metricsRouter := chi.NewRouter()

		metricsRouter.Handle("/metrics", promhttp.Handler())
		metricsRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		log.Printf("[payment-service] metrics server on :%s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, metricsRouter); err != nil {
			log.Fatal(err)
		}
	}()

	appRouter := chi.NewRouter()

	// RED metrics for all application routes.
	appRouter.Use(httpx.Middleware(redMetrics))

	appRouter.Post("/pay", srv.handlePay)
	appRouter.Get("/health", srv.handleHealth)

	log.Printf("[payment-service] HTTP server on :%s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, appRouter); err != nil {
		log.Fatal(err)
	}
}
