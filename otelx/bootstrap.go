// Package otelx provides one shared way for every shop_* service to wire up
// OpenTelemetry: OTLP/gRPC exporters for traces and metrics, resource
// attributes, and a single Shutdown func. See
// shop_docs/docs/observability.md.
package otelx

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config controls what a service's telemetry identifies itself as and
// where it's shipped. Endpoint is normally the in-cluster OTel Collector
// (see docs/observability.md), not a vendor endpoint directly.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Endpoint       string // e.g. "otel-collector.observability.svc.cluster.local:4317"
	Insecure       bool   // true for in-cluster plaintext gRPC (no public exposure)
}

// Shutdown flushes and closes every provider this package set up. Callers
// should defer it (with a bounded-timeout context) right after Bootstrap
// succeeds.
type Shutdown func(ctx context.Context) error

// Bootstrap wires up global TracerProvider and MeterProvider instances and
// the W3C trace-context + baggage propagators. Every shop_* service calls
// this once at startup.
func Bootstrap(ctx context.Context, cfg Config) (Shutdown, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otelx: build resource: %w", err)
	}

	dialOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
	metricDialOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		dialOpts = append(dialOpts, otlptracegrpc.WithInsecure())
		metricDialOpts = append(metricDialOpts, otlpmetricgrpc.WithInsecure())
	}

	traceExporter, err := otlptracegrpc.New(ctx, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("otelx: create trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	metricExporter, err := otlpmetricgrpc.New(ctx, metricDialOpts...)
	if err != nil {
		return nil, fmt.Errorf("otelx: create metric exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(15*time.Second))),
	)
	otel.SetMeterProvider(meterProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("otelx: shutdown tracer provider: %w", err)
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("otelx: shutdown meter provider: %w", err)
		}
		return nil
	}, nil
}
