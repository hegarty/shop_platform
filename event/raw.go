// Package event defines the platform's event envelopes: the raw ingestion
// envelope wrapping a source system's native payload, and the canonical
// commerce.* event normalized from it. Downstream consumers should depend
// on the canonical shape, not on Shopify's (or any other source's) native
// schema — see shop_docs/docs/event-model.md.
package event

import (
	"encoding/json"
	"errors"
	"time"
)

// RawEnvelope wraps a source system's native webhook/API payload before
// normalization. Published to topics like shopify.orders.raw. Preserves
// enough metadata (source, shop domain, receipt time) to investigate
// classification errors without needing to replay from Shopify itself.
type RawEnvelope struct {
	EventID       string          `json:"event_id"`
	TenantID      string          `json:"tenant_id"`
	Source        string          `json:"source"`
	EventType     string          `json:"event_type"`
	ShopDomain    string          `json:"shop_domain"`
	ReceivedAt    time.Time       `json:"received_at"`
	SchemaVersion int             `json:"schema_version"`
	Payload       json.RawMessage `json:"payload"`
}

// CurrentRawSchemaVersion is the schema_version stamped on newly-created
// RawEnvelopes. Bump when the envelope shape changes in a way consumers
// need to branch on.
const CurrentRawSchemaVersion = 1

var (
	ErrMissingEventID    = errors.New("event: missing event_id")
	ErrMissingTenantID   = errors.New("event: missing tenant_id")
	ErrMissingSource     = errors.New("event: missing source")
	ErrMissingEventType  = errors.New("event: missing event_type")
	ErrMissingPayload    = errors.New("event: missing payload")
	ErrMissingID         = errors.New("event: missing id")
	ErrMissingType       = errors.New("event: missing type")
	ErrMissingAttributes = errors.New("event: missing attributes")
)

// Validate checks that the envelope has the minimum fields required to be
// published and later traced back to its source.
func (e RawEnvelope) Validate() error {
	switch {
	case e.EventID == "":
		return ErrMissingEventID
	case e.TenantID == "":
		return ErrMissingTenantID
	case e.Source == "":
		return ErrMissingSource
	case e.EventType == "":
		return ErrMissingEventType
	case len(e.Payload) == 0:
		return ErrMissingPayload
	}
	return nil
}
