package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/Ley-code/Testing_Task_3/internal/config"
	"github.com/Ley-code/Testing_Task_3/observability"
)

// Default OpenTelemetry global provider is a no-op until you register an OTLP/stdout SDK
// (see TRACING.md). Handlers still call StartSpan so wiring matches production.

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("warning: load .env: %v", err)
	}
	logger := observability.NewLogger(nil)
	metrics := observability.NewOrderMetrics("order")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := observability.StartSpan(ctx, "demo-handler")
		defer span.End()

		metrics.IncCreated("ok", "standard")
		stepStart := time.Now()
		metrics.ObserveStep("validation", time.Since(stepStart))
		logger.Info(ctx, "demo request", logrus.Fields{"path": r.URL.Path})

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("demo ok\n"))
	})

	addr := config.ListenAddr()
	log.Printf("listening on %s (GET /health /metrics /demo)", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
