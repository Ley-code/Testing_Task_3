# Logging strategy — Order Processing Service

## Where should logging happen?

### 1. Which option maintains Clean Architecture?

- **Operational / request lifecycle logs:** Prefer **Option B** (service/transport boundary) or **Option C** (decorator around the usecase interface): one place for “request in / outcome / latency / error,” keeps usecases free of repetitive boilerplate.
- **Business / audit logs:** **Option A** (inside usecase) is appropriate when the message is a **deliberate business fact** (e.g. “order accepted for fulfillment”) using **safe, structured fields** — not raw request payloads.
- **Domain:** No logging in entities/value objects; domain remains pure.

Combining **B/C for operations** and **targeted A for audit** preserves boundaries: domain silent, adapters and application policy explicit.

### 2. Business logs vs operational logs

| Type | Purpose | Examples |
|------|---------|----------|
| **Business** | Product, compliance, audit trails | Order placed, refund initiated, tier upgrade |
| **Operational** | SRE/debug: health, latency, retries, dependency errors | HTTP 502 from Payment, DB pool exhausted, handler panic |

Business logs are often retained longer and may be subject to policy; operational logs power dashboards and on-call.

### 3. Avoiding sensitive data

- **Structured fields only** at boundaries; never `log.Info("req", req)` for arbitrary structs.
- **Allowlisted keys** for debug; **denylist/redact** keys matching `password`, `token`, `authorization`, `pan`, `cvv`, etc.
- **Truncate** large bodies; use correlation IDs instead of full PAN or card numbers.

## Implementation (this repo)

[`observability/logging.go`](observability/logging.go) implements:

- **Trace ID in every log:** `Logger` builds a `logrus.Entry` from `trace.SpanFromContext(ctx)` and adds `trace_id` and `span_id` when the span is valid.
- **Request/response at boundaries:** `LogRequest` / `LogResponse` with operation name and sanitized maps (via `RedactMap`).
- **Errors with stack traces:** `Error(..., withStack bool)` attaches `runtime/debug.Stack()` at **boundaries** only (not inside domain).
- **PII redaction:** `RedactMap` masks sensitive keys; `RedactString` is a best-effort scrub for free-form text.

**Example at HTTP boundary:**

```go
obs.LogRequest(ctx, "POST /orders", map[string]string{"customer_id": cid, "payment_token": tok})
// payment_token -> [REDACTED]
```
