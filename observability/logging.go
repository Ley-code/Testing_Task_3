package observability

import (
	"context"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

var sensitiveKeyPattern = regexp.MustCompile(`(?i)(password|passwd|secret|token|authorization|cvv|pan|card)`)

// Logger wraps logrus with trace correlation and redaction helpers.
type Logger struct {
	inner *logrus.Logger
}

// NewLogger builds a logger; if inner is nil, a default JSON formatter is used.
func NewLogger(inner *logrus.Logger) *Logger {
	if inner == nil {
		inner = logrus.New()
		inner.SetFormatter(&logrus.JSONFormatter{})
	}
	return &Logger{inner: inner}
}

func (l *Logger) entry(ctx context.Context) *logrus.Entry {
	e := logrus.NewEntry(l.inner)
	if ctx == nil {
		return e
	}
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return e
	}
	sc := span.SpanContext()
	return e.WithFields(logrus.Fields{
		"trace_id": sc.TraceID().String(),
		"span_id":  sc.SpanID().String(),
	})
}

// Info logs at info with trace fields.
func (l *Logger) Info(ctx context.Context, msg string, fields logrus.Fields) {
	l.entry(ctx).WithFields(fields).Info(msg)
}

// Error logs an error with optional stack at operational boundaries.
func (l *Logger) Error(ctx context.Context, msg string, err error, withStack bool, fields logrus.Fields) {
	e := l.entry(ctx).WithFields(fields)
	if err != nil {
		e = e.WithError(err)
	}
	if withStack {
		e = e.WithField("stack", string(debug.Stack()))
	}
	e.Error(msg)
}

// LogRequest logs a sanitized snapshot at an API/service boundary.
func (l *Logger) LogRequest(ctx context.Context, operation string, fields map[string]string) {
	safe := RedactMap(fields)
	l.entry(ctx).WithFields(logrus.Fields{
		"operation": operation,
		"fields":    safe,
	}).Info("request")
}

// LogResponse logs a coarse outcome at a boundary.
func (l *Logger) LogResponse(ctx context.Context, operation string, statusCode int, err error) {
	e := l.entry(ctx).WithFields(logrus.Fields{
		"operation":   operation,
		"status_code": statusCode,
	})
	if err != nil {
		e = e.WithError(err)
	}
	e.Info("response")
}

// RedactMap returns a copy with sensitive keys masked.
func RedactMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		if sensitiveKeyPattern.MatchString(k) {
			out[k] = "[REDACTED]"
			continue
		}
		out[k] = v
	}
	return out
}

// RedactString masks common secret patterns in free-form strings (best-effort).
func RedactString(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Fields(s)
	for i, p := range parts {
		if len(p) > 20 && strings.Contains(p, ".") {
			parts[i] = "[REDACTED]"
		}
	}
	return strings.Join(parts, " ")
}
