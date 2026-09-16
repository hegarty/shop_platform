package redpanda

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Handler processes one record. A returned error does NOT stop the
// consumer or block committing later records — retry/dead-letter policy
// (e.g. publishing to a *.failed topic after N attempts) is the caller's
// responsibility, not this package's. See shop_docs/docs/event-model.md's
// error-handling section.
type Handler func(ctx context.Context, record *kgo.Record) error

// Consumer polls a consumer group and commits offsets after each record's
// Handler returns (whether it succeeded or not — errors are the caller's
// concern, per Handler's contract).
type Consumer struct {
	client *kgo.Client
}

// NewConsumer connects to the given brokers as a member of group, consuming
// topics. Offsets are committed manually (via CommitRecords), one record's
// worth at a time, rather than franz-go's default auto-commit — this keeps
// "processed" and "committed" tightly coupled instead of committing
// slightly ahead of what Handler has actually finished.
func NewConsumer(brokers []string, group string, topics []string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
	)
	if err != nil {
		return nil, fmt.Errorf("redpanda: new consumer client: %w", err)
	}
	return &Consumer{client: client}, nil
}

// Run polls in a loop, invoking handler for every record, until ctx is
// cancelled or a poll returns a fatal client error.
func (c *Consumer) Run(ctx context.Context, handler Handler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			c.client.AllowRebalance()
			return fmt.Errorf("redpanda: poll fetch error: %v", errs[0].Err)
		}

		fetches.EachRecord(func(record *kgo.Record) {
			c.processOne(ctx, record, handler)
		})

		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			c.client.AllowRebalance()
			return fmt.Errorf("redpanda: commit offsets: %w", err)
		}
		c.client.AllowRebalance()
	}
}

func (c *Consumer) processOne(ctx context.Context, record *kgo.Record, handler Handler) {
	headers := record.Headers
	spanCtx := otel.GetTextMapPropagator().Extract(ctx, headerCarrier{headers: &headers})

	spanCtx, span := otel.Tracer(tracerName).Start(spanCtx, "redpanda.consume",
		trace.WithAttributes(
			attribute.String("messaging.system", "redpanda"),
			attribute.String("messaging.destination", record.Topic),
			attribute.Int64("messaging.kafka.partition", int64(record.Partition)),
			attribute.Int64("messaging.kafka.offset", record.Offset),
		),
	)
	defer span.End()

	if err := handler(spanCtx, record); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "handler failed")
	}
}

// Close releases broker connections without committing further offsets.
func (c *Consumer) Close() {
	c.client.Close()
}
