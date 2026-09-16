// Package logging provides one structured (JSON) logger setup shared by
// every shop_* service, with trace/span IDs automatically attached to every
// log line that's emitted with a context carrying an active OpenTelemetry
// span — see shop_docs/docs/observability.md's correlation requirements.
package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

// New builds a JSON slog.Logger at the given level, writing to os.Stdout
// (container runtimes capture stdout; this platform doesn't write log
// files). service is attached to every line so multi-service log queries
// (Loki) can filter by it.
func New(service string, level slog.Level) *slog.Logger {
	handler := &traceHandler{
		inner: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
	}
	return slog.New(handler).With(slog.String("service", service))
}

// traceHandler wraps another slog.Handler, adding trace_id/span_id
// attributes from ctx's active span, when present.
type traceHandler struct {
	inner slog.Handler
}

func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *traceHandler) Handle(ctx context.Context, record slog.Record) error {
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", span.TraceID().String()),
			slog.String("span_id", span.SpanID().String()),
		)
	}
	return h.inner.Handle(ctx, record)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{inner: h.inner.WithGroup(name)}
}

// sensitiveHeaders is used by RedactHeader to decide which HTTP header
// values must never reach a log line, even at debug level.
var sensitiveHeaders = map[string]bool{
	"authorization":          true,
	"x-shopify-hmac-sha256":  true,
	"x-shopify-access-token": true,
	"cookie":                 true,
	"set-cookie":             true,
	"x-api-key":              true,
}

// RedactHeader returns value unchanged unless key is a known
// credential-carrying header (case-insensitive), in which case it returns
// a fixed placeholder. Webhook receivers and outbound HTTP clients should
// route every header through this before logging a request.
func RedactHeader(key, value string) string {
	if sensitiveHeaders[strings.ToLower(key)] {
		return "[REDACTED]"
	}
	return value
}
