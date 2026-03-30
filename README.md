# Testing Task 3 — Observability & integration testing

Small Go workspace demonstrating structured logging, Prometheus metrics, OpenTelemetry tracing helpers, and Testcontainers-based integration tests.

## Layout

| Path | Purpose |
|------|---------|
| `cmd/api` | HTTP entrypoint: `/health`, `/metrics`, `/demo` (uses `observability`) |
| `internal/config` | Minimal config (e.g. `PORT` for the server) |
| `observability` | Logging, metrics, tracing helpers |
| `testing/setup` | Postgres + WireMock testcontainers |
| `testing/integration` | Integration tests (Docker required) |
| `*.md` | Task answers and notes (`ANSWERS.md`, `LOGGING.md`, …) |

## Run the API

```bash
go run ./cmd/api
```

Optional: `PORT=3000 go run ./cmd/api` — then open `http://localhost:3000/health`, `http://localhost:3000/metrics`, `http://localhost:3000/demo`.

## Tests

```bash
go test ./... -short
```

Fast path; skips integration tests.

```bash
go test ./... 
```

Runs integration tests (needs Docker). If Docker is unavailable, those tests skip with a clear message.

## Requirements

- Go 1.18+
- Docker (for integration tests only)

## GitHub repository metadata (optional)

Suggested **description**: *Go observability patterns: logging, Prometheus metrics, OpenTelemetry, Testcontainers integration tests.*

Suggested **topics**: `go`, `opentelemetry`, `prometheus`, `testcontainers`, `integration-testing`, `observability`

## Incremental commits

Small, reviewable commits work well, for example:

1. `chore: module path, gitignore`
2. `feat: cmd/api and internal/config`
3. `docs: README`

Repeat that pattern as you change `observability/` or tests.
