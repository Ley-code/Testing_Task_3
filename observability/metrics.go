package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// OrderMetrics holds Prometheus instruments with bounded label sets only.
type OrderMetrics struct {
	OrdersCreated          *prometheus.CounterVec
	OrderProcessingSeconds *prometheus.HistogramVec
	OrdersPending          prometheus.Gauge
}

// NewOrderMetrics registers counters/histograms/gauges for order processing.
// Labels are intentionally low-cardinality (no customer_id, order_id, product_id).
func NewOrderMetrics(namespace string) *OrderMetrics {
	if namespace == "" {
		namespace = "order"
	}
	return &OrderMetrics{
		OrdersCreated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "orders_created_total",
				Help:      "Orders created by outcome and customer tier.",
			},
			[]string{"status", "customer_tier"},
		),
		OrderProcessingSeconds: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "order_processing_duration_seconds",
				Help:      "Time spent in pipeline steps (validation, payment, fulfillment).",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"step"},
		),
		OrdersPending: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "orders_pending_count",
				Help:      "Current number of orders awaiting completion.",
			},
		),
	}
}

// ObserveStep records duration for a pipeline step (use exemplars in scrape config / Grafana for trace correlation).
func (m *OrderMetrics) ObserveStep(step string, d time.Duration) {
	if m == nil || m.OrderProcessingSeconds == nil {
		return
	}
	m.OrderProcessingSeconds.WithLabelValues(step).Observe(d.Seconds())
}

// IncCreated increments order creation counter.
func (m *OrderMetrics) IncCreated(status, customerTier string) {
	if m == nil || m.OrdersCreated == nil {
		return
	}
	m.OrdersCreated.WithLabelValues(status, customerTier).Inc()
}

// SetPending sets the pending orders gauge.
func (m *OrderMetrics) SetPending(n float64) {
	if m == nil || m.OrdersPending == nil {
		return
	}
	m.OrdersPending.Set(n)
}
