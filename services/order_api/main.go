package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"shared/chaos"
	"shared/httpx"
	"shared/utils"
	"strconv"
)

func main() {
	httpPort := utils.EnvOrDefault("HTTP_PORT", "8081")
	metricsPort := utils.EnvOrDefault("METRICS_PORT", "9091")
	processorURL := utils.EnvOrDefault("ORDER_PROCESSOR_URL", "http://localhost:8082")

	admissionMaxStr := utils.EnvOrDefault("ORDER_API_ADMISSION_MAX_IN_FLIGHT", "100")
	admissionMax, err := strconv.Atoi(admissionMaxStr)
	if err != nil || admissionMax <= 0 {
		log.Fatalf("[order-api] invalid ORDER_API_ADMISSION_MAX_IN_FLIGHT=%q: must be a positive integer", os.Getenv("ORDER_API_ADMISSION_MAX_IN_FLIGHT"))
	}

	reg := prometheus.DefaultRegisterer

	serviceMetrics := NewMetrics(reg)
	redMetrics := httpx.NewREDMetrics(reg, "order_api")
	chaosMetrics := chaos.NewMetrics(reg, "order_api")

	chaosCfg := chaos.NewConfig("ORDER_API")

	admission := newSimpleAdmissionControl(admissionMax)
	serviceMetrics.AdmissionLimit.Set(float64(admissionMax))

	srv := NewServer(processorURL, serviceMetrics, admission)

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

	log.Printf("[order-api] HTTP server on :%s | processor=%s | admission_limit=%d", httpPort, processorURL, admissionMax)
	if err := http.ListenAndServe(":"+httpPort, srv.routes(redMetrics, chaosCfg, chaosMetrics)); err != nil {
		log.Fatal(err)
	}
}
