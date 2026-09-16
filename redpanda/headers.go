package redpanda

import "github.com/twmb/franz-go/pkg/kgo"

// headerCarrier adapts a *[]kgo.RecordHeader to OpenTelemetry's
// propagation.TextMapCarrier, so trace context can ride along in Kafka/
// Redpanda record headers across the ingestor -> analytics -> notifier
// hop, per shop_docs/docs/observability.md.
type headerCarrier struct {
	headers *[]kgo.RecordHeader
}

func (h headerCarrier) Get(key string) string {
	for _, hd := range *h.headers {
		if hd.Key == key {
			return string(hd.Value)
		}
	}
	return ""
}

func (h headerCarrier) Set(key, value string) {
	for i, hd := range *h.headers {
		if hd.Key == key {
			(*h.headers)[i].Value = []byte(value)
			return
		}
	}
	*h.headers = append(*h.headers, kgo.RecordHeader{Key: key, Value: []byte(value)})
}

func (h headerCarrier) Keys() []string {
	keys := make([]string, len(*h.headers))
	for i, hd := range *h.headers {
		keys[i] = hd.Key
	}
	return keys
}
