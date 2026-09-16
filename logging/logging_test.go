package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestRedactHeader(t *testing.T) {
	cases := []struct {
		key, value, want string
	}{
		{"Authorization", "Bearer secret", "[REDACTED]"},
		{"X-Shopify-Hmac-Sha256", "abc123", "[REDACTED]"},
		{"x-shopify-access-token", "shpat_xxx", "[REDACTED]"},
		{"Content-Type", "application/json", "application/json"},
	}
	for _, c := range cases {
		if got := RedactHeader(c.key, c.value); got != c.want {
			t.Errorf("RedactHeader(%q, %q) = %q, want %q", c.key, c.value, got, c.want)
		}
	}
}

func TestNew_AttachesTraceIDFromContext(t *testing.T) {
	var buf bytes.Buffer
	handler := &traceHandler{inner: slog.NewJSONHandler(&buf, nil)}
	logger := slog.New(handler)

	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	logger.InfoContext(ctx, "test message")

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal log line: %v", err)
	}
	if decoded["trace_id"] != traceID.String() {
		t.Errorf("trace_id = %v, want %v", decoded["trace_id"], traceID.String())
	}
	if decoded["span_id"] != spanID.String() {
		t.Errorf("span_id = %v, want %v", decoded["span_id"], spanID.String())
	}
}

func TestNew_NoSpan_NoTraceFields(t *testing.T) {
	var buf bytes.Buffer
	handler := &traceHandler{inner: slog.NewJSONHandler(&buf, nil)}
	logger := slog.New(handler)

	logger.InfoContext(context.Background(), "no span here")

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal log line: %v", err)
	}
	if _, ok := decoded["trace_id"]; ok {
		t.Error("expected no trace_id field without an active span")
	}
}
