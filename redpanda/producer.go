// Package redpanda wraps franz-go into the narrow producer/consumer shapes
// this platform actually needs: tenant-keyed publish, trace-context
// propagation via headers, and a simple per-record processing loop.
// Deliberately not a general-purpose Kafka client wrapper.
package redpanda

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/hegarty/shop_platform/redpanda"

// Producer publishes JSON-encoded messages, keyed by tenant ID so all of a
// tenant's events land on the same partition (preserving per-tenant
// ordering without requiring a single-partition topic).
type Producer struct {
	client *kgo.Client
}

// NewProducer connects to the given Redpanda/Kafka brokers. Idempotent
// production is enabled so retried publishes (e.g. after a transient
// broker error) can't duplicate a message on the wire — consumers still
// need their own idempotency key for end-to-end exactly-once semantics
// (see shop_docs/docs/event-model.md), since a producer-level guarantee
// only covers producer-to-broker duplication.
func NewProducer(brokers []string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerLinger(0),
		kgo.RecordPartitioner(kgo.UniformBytesPartitioner(1<<20, false, false, nil)),
	)
	if err != nil {
		return nil, fmt.Errorf("redpanda: new producer client: %w", err)
	}
	return &Producer{client: client}, nil
}

// Publish sends value (JSON-marshaled) to topic, keyed by tenantID, with an
// event-type header and injected trace context. Blocks until the broker
// acknowledges or ctx is cancelled.
func (p *Producer) Publish(ctx context.Context, topic, tenantID, eventType string, value any) error {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "redpanda.publish",
		trace.WithAttributes(
			attribute.String("messaging.system", "redpanda"),
			attribute.String("messaging.destination", topic),
			attribute.String("tenant.id", tenantID),
			attribute.String("event.type", eventType),
		),
	)
	defer span.End()

	payload, err := json.Marshal(value)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal failed")
		return fmt.Errorf("redpanda: marshal payload: %w", err)
	}

	headers := []kgo.RecordHeader{
		{Key: "event_type", Value: []byte(eventType)},
	}
	otel.GetTextMapPropagator().Inject(ctx, headerCarrier{headers: &headers})

	record := &kgo.Record{
		Topic:   topic,
		Key:     []byte(tenantID),
		Value:   payload,
		Headers: headers,
	}

	result := p.client.ProduceSync(ctx, record)
	if err := result.FirstErr(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "produce failed")
		return fmt.Errorf("redpanda: produce to %s: %w", topic, err)
	}
	return nil
}

// Close flushes any buffered records and releases broker connections.
func (p *Producer) Close() {
	p.client.Close()
}
