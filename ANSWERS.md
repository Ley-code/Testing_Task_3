# Assessment answers — Q1–Q5

## Q1 — Trace domain operations without passing `context` (two approaches)

**Problem:** Putting `trace.SpanFromContext(ctx)` inside `Order.Complete` couples domain to OpenTelemetry and breaks purity.

**Approach 1 — Span around pure domain calls (application/usecase):**  
The usecase starts a span (e.g. `domain.Order.Complete`) immediately before `order.Complete()` and ends it when the method returns. The domain method stays `(o *Order) Complete() error` with no `context`.

**Approach 2 — Domain events / results:**  
The domain returns facts or emits domain events (`OrderCompleted`). The usecase subscribes to those outcomes **outside** the domain and adds span events or child spans. Tracing stays in the application layer; domain only knows business rules.

---

## Q2 — Mocking: when it helps, when it hides bugs, when you need both

### 1. When mocking is the **right** choice

- **Pure logic** with no I/O: e.g. pricing rules, discount caps, state machine transitions — mock **time** or **repo** to fix inputs and assert outputs quickly.
- **Isolating failure modes** of a collaborator: simulate “payment declined” without calling a real PSP.

### 2. When mocking **hides** a bug (example)

- **Bug:** `UPDATE orders SET status = $1` uses the wrong placeholder order, so status is written to the wrong column, or the statement is never committed.
- **Mock:** `MockRepo.On("Save").Return(nil)` always succeeds — the test passes while production corrupts data.

### 3. When you need **both** mock tests **and** integration tests

- **Unit (mock):** Fast feedback on domain rules and usecase branching when the DB is irrelevant.
- **Integration (real Postgres):** Prove SQL, transactions, and constraints match reality.  
Together: correctness of logic **and** persistence.

---

## Q3 — “Order was charged but shows as failed” — debugging workflow

### Logs that help

- Single **`trace_id`** across Order Service and Payment Service requests.
- **Payment:** charge id, amount, PSP response code, idempotency key.
- **Order:** state transitions (`pending` → `charging` → `paid` / `failed`), who wrote the final status, error from downstream.
- **Outbox / async:** if status is updated by a consumer, log publish and ack.

### Metrics that indicate this

- **`orders_created_total{status="failure"}`** vs **`payment_charges_total{status="success"}`** (or PSP success counter) **diverging** in the same window.
- **`order_processing_duration_seconds{step="payment"}`** latency vs errors.
- Optional **business counter** `orders_charge_mismatch_total` incremented when payment reports success but order update fails (detected in compensating logic).

### Alerts

- Alert on **rate of “charged but order not terminal success”** (custom metric or log-based alert).
- **SLO burn** on successful order completion after payment callback.
- **Spike** in `payment` step errors or **drop** in inventory fulfillment after charge.

---

## Q4 — Outbox pattern E2E **without** flaky timing

Avoid blind `time.Sleep` waiting for the worker.

1. **Same database transaction:** Persist aggregate + outbox row in **one commit** — test can query `outbox` immediately after commit and see the row.
2. **Deterministic worker trigger:** In tests, call the worker’s **ProcessBatch** (or poll **once**) **synchronously** from the test after insert, or use an **in-process** queue with **blocking** receive instead of real time delays.
3. **Poll with context deadline:** If async is required, `for { select { case <-ctx.Done(): ... default: if processed() { break } } }` with a **short** timeout instead of fixed sleep.
4. **Fake clock** only where the worker schedules retries — advance clock explicitly in tests.

Use **Testcontainers** for Postgres + a **test double** for the message broker or an embedded broker with synchronous subscription.

---

## Q5 — Test passes locally, fails in CI: `time.Now()` comparison

**What’s wrong:** The test sets `order.CreatedAt = time.Now()`, then later asserts `order.CreatedAt` equals **another** `time.Now()`. Wall clock moves between the two calls, so equality is **nondeterministic** (worse under CI load).

**Fix:**

- Inject a **`Clock`** interface (`Now() time.Time`) or a **`func() time.Time`** into code under test; tests pass a **fixed** instant.
- Or assert **approximate** equality: `assert.WithinDuration(t, want, got, time.Second)` with a single captured `now := time.Now()` used for both setup and expectation when appropriate.

Never assert exact `time.Now()` equality across non-atomic statements.
