package main

import (
	"log"
	"net/http"
	"shared/chaos"
	"shared/httpx"
	"shared/utils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	httpPort := utils.EnvOrDefault("HTTP_PORT", "8081")
	metricsPort := utils.EnvOrDefault("METRICS_PORT", "9091")
	processorURL := utils.EnvOrDefault("ORDER_PROCESSOR_URL", "http://localhost:8082")

	reg := prometheus.DefaultRegisterer

	serviceMetrics := NewMetrics(reg)
	redMetrics := httpx.NewREDMetrics(reg, "order_api")
	chaosMetrics := chaos.NewMetrics(reg, "order_api")

	chaosCfg := chaos.NewConfig("ORDER_API")
	srv := NewServer(processorURL, serviceMetrics)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		log.Printf("[order-api] metrics server on :%s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("[order-api] HTTP server on :%s | processor=%s", httpPort, processorURL)
	if err := http.ListenAndServe(":"+httpPort, srv.routes(redMetrics, chaosCfg, chaosMetrics)); err != nil {
		log.Fatal(err)
	}
}
