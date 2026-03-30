# Metrics design — Order Processing Service

## Instrument definitions

```text
# Counters
orders_created_total{status="success|failure", customer_tier="free|premium"}

# Histograms
order_processing_duration_seconds{step="validation|payment|fulfillment"}

# Gauges
orders_pending_count{}
```

Implementation: [`observability/metrics.go`](observability/metrics.go) — `NewOrderMetrics`, `IncCreated`, `ObserveStep`, `SetPending`.

## 1. Where do you instrument?

- **Primary:** **Usecase / application service** boundary — one place per business operation avoids double-counting and matches user-perceived latency (validation → payment → fulfillment).
- **Not** inside pure domain entities.
- **Repository (optional):** repo-level histograms for DB latency are possible for SLO splits, but label carefully; prefer a single “step=fulfillment” bucket unless you need granular DB tuning.

## 2. Metric explosion — example and fix

**Bad:**

```go
orderDuration.WithLabelValues(customerID, productID, orderID).Observe(duration)
```

**Why it explodes:** `customer_id`, `product_id`, and `order_id` are **high-cardinality** labels. Prometheus creates a new time series per distinct label combination → memory blow-up, slow queries, useless aggregation.

**Fix:**

- Use **bounded** labels: `status`, `customer_tier`, `step`, `error_code` (enum), `deployment`, `region`.
- Put **order_id** / **trace_id** in **logs and traces**, not metric labels.
- Use **exemplars** (where supported) on histograms to attach a **trace ID** to individual observations for drill-down without labeling every series.

## 3. Correlating metrics with traces

- Same **`trace_id`** in structured logs and span context.
- **OpenTelemetry resource** attributes: `service.name`, `service.version`, `deployment.environment` — align with Prometheus `external_labels` or Grafana datasources.
- **Exemplars** on `order_processing_duration_seconds` linking to `trace_id` (Prometheus 2.26+, Grafana/Mimir exemplar views).
- Dashboards: jump from spike in error rate → exemplar → trace → logs.
