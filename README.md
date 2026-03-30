# Testing Task 3 — Observability & integration testing

Small Go workspace demonstrating structured logging, Prometheus metrics, OpenTelemetry tracing helpers, and Testcontainers-based integration tests.

## Layout

| Path | Purpose |
|------|---------|
| `cmd/api` | HTTP entrypoint: `/health`, `/metrics`, `/demo` (uses `observability`) |
| `internal/config` | Minimal config (e.g. `PORT` for the server) |
| `.env.example` | Documented environment variables; copy to `.env` for local runs |
| `observability` | Logging, metrics, tracing helpers |
| `testing/setup` | Postgres + WireMock testcontainers |
| `testing/integration` | Integration tests (Docker required) |
| `*.md` | Task answers and notes (`ANSWERS.md`, `LOGGING.md`, …) |

## Run the API

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port (`8080` or `:8080`) |

Copy the example file and edit values as needed:

```bash
cp .env.example .env
```

The process loads `.env` from the current working directory when the file exists (ignored by git). You can still override with `PORT=3000 go run ./cmd/api`.

### Locally (Go installed)

```bash
go run ./cmd/api
```

With a custom port without a file: `PORT=3000 go run ./cmd/api`.

### With Docker

Build the image:

```bash
docker build -t order-api .
```

Run the container:

```bash
docker run --rm -p 8080:8080 order-api
```

Or override the port:

```bash
docker run --rm -e PORT=3000 -p 3000:3000 order-api
```

Or pass a file (e.g. after `cp .env.example .env`):

```bash
docker run --rm --env-file .env -p 8080:8080 order-api
```

### Verify

Once the server is running, open any of these endpoints:

- `http://localhost:8080/health`  — returns `ok`
- `http://localhost:8080/metrics` — Prometheus metrics
- `http://localhost:8080/demo`    — exercises logging, metrics, and tracing

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
- Docker (for running the container and integration tests)

## GitHub repository metadata (optional)

Suggested **description**: *Go observability patterns: logging, Prometheus metrics, OpenTelemetry, Testcontainers integration tests.*

Suggested **topics**: `go`, `opentelemetry`, `prometheus`, `testcontainers`, `integration-testing`, `observability`

## Incremental commits

Small, reviewable commits work well, for example:

1. `chore: module path, gitignore`
2. `feat: cmd/api and internal/config`
3. `docs: README`

Repeat that pattern as you change `observability/` or tests.
