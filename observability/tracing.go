package observability

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider returns the global tracer provider (set by app bootstrap).
func TracerProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

// Tracer returns a tracer for the order processing service.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// SpanFromContext returns the current span (may be noop).
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// StartSpan starts a child span under the current context.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer("order-processing").Start(ctx, name, opts...)
}

// InjectHTTP injects trace context into outgoing HTTP headers (W3C Trace Context).
func InjectHTTP(ctx context.Context, header http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
}

// ExtractHTTP returns a context with remote span context from incoming HTTP headers.
func ExtractHTTP(ctx context.Context, header http.Header) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

// RunInSpan starts a child span, runs fn, and ends the span. Use from usecases to wrap
// calls into pure domain code (domain stays free of context/trace imports).
func RunInSpan(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := Tracer("order-processing").Start(ctx, name)
	defer span.End()
	return fn(ctx)
}
