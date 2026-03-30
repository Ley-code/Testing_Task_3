# Integration testing strategy — Order Processing Service

## The over-mocking problem

Unit tests that mock the repository to always succeed **hide**:

1. Invalid or dialect-wrong **SQL** (mock never hits the database).
2. **Serialization** bugs (JSON/tags vs column types).
3. **Database constraints** (unique keys, FKs, check constraints, `NOT NULL`).

The assessment example passes without exercising real persistence.

## Test pyramid

| Test level | What to test | What to mock / fake |
|------------|----------------|---------------------|
| **Unit** | Pure domain rules, value objects, pricing/invariants | Clock (`time.Now`), repos, HTTP clients, message buses |
| **Integration** | Real SQL, migrations, type mapping, constraint behavior; HTTP client against **WireMock** or test double | Nothing inside the DB container; external systems replaced by **Testcontainers** (Postgres, WireMock) |
| **E2E** | Full path: API → Order → DB → Payment/Inventory (test env) | Only third parties you cannot run locally (or use contract tests + sandboxes) |

## Implementations in this repo

### 1. Repository integration test (Testcontainers Postgres)

- [`testing/setup/testcontainers.go`](testing/setup/testcontainers.go) — `StartPostgres` starts `postgres:15-alpine`, returns a `postgres://` DSN.
- [`testing/integration/order_repo_test.go`](testing/integration/order_repo_test.go) — `TestOrderRepo_Integration` creates schema, inserts via `CreateMut`, asserts row content, updates status via `UpdateMut` with a **dirty-field map**, verifies DB state.

**Running:** requires Docker. Skipped with `go test -short`.

```bash
go test ./testing/integration/... -count=1
```

### 2. External service test (WireMock)

- [`testing/setup/testcontainers.go`](testing/setup/testcontainers.go) — `StartWireMock` starts `wiremock/wiremock:3.3.1`.
- [`testing/integration/payment_client_test.go`](testing/integration/payment_client_test.go) — registers stubs via Admin API; tests **success JSON**, **HTTP 500**, and **client timeout** (unroutable address + short timeout).

**CI without Docker:** integration tests **skip** if Testcontainers cannot reach Docker (with a clear `Skipf` reason). Use `go test -short` in fast pipelines to skip containers explicitly.

## Design choices

- **Short mode:** `testing.Short()` skips container-based tests for quick feedback.
- **No flaky sleeps:** WireMock readiness uses `wait.ForListeningPort`; assertions use HTTP responses, not arbitrary `time.Sleep` for “eventually.”
