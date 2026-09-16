package event

import (
	"encoding/json"
	"time"
)

// Type is a canonical platform event type, independent of the source
// system that produced it. Consumers should switch on these, never on a
// source-specific webhook topic string.
type Type string

const (
	TypeOrderCreated      Type = "commerce.order.created"
	TypeOrderUpdated      Type = "commerce.order.updated"
	TypeOrderRefunded     Type = "commerce.order.refunded"
	TypeInventoryChanged  Type = "inventory.changed"
	TypeCustomerCreated   Type = "customer.created"
	TypeShipmentCreated   Type = "shipment.created"
	TypeShipmentDelivered Type = "shipment.delivered"
	TypeAdSpendRecorded   Type = "ad.spend.recorded"
)

// CurrentCommerceSchemaVersion is the version stamped on newly-normalized
// CommerceEvents.
const CurrentCommerceSchemaVersion = 1

// CommerceEvent is the canonical, source-agnostic event published after
// normalization (e.g. to commerce.events or shopify.orders.normalized).
// Attributes is intentionally typed per-Type (OrderAttributes for the
// commerce.order.* family today) rather than a single giant struct, so
// adding a new event family doesn't touch existing ones.
type CommerceEvent struct {
	ID         string          `json:"id"`
	TenantID   string          `json:"tenant_id"`
	Type       Type            `json:"type"`
	Source     string          `json:"source"`
	Timestamp  time.Time       `json:"timestamp"`
	Version    int             `json:"version"`
	Attributes json.RawMessage `json:"attributes"`
}

// Validate checks the minimum fields required for a CommerceEvent to be
// meaningfully processed downstream.
func (c CommerceEvent) Validate() error {
	switch {
	case c.ID == "":
		return ErrMissingID
	case c.TenantID == "":
		return ErrMissingTenantID
	case c.Type == "":
		return ErrMissingType
	case len(c.Attributes) == 0:
		return ErrMissingAttributes
	}
	return nil
}

// DecodeOrderAttributes unmarshals Attributes into an OrderAttributes
// struct. Callers should check Type is one of the commerce.order.* family
// before calling this.
func (c CommerceEvent) DecodeOrderAttributes() (OrderAttributes, error) {
	var attrs OrderAttributes
	if err := json.Unmarshal(c.Attributes, &attrs); err != nil {
		return OrderAttributes{}, err
	}
	return attrs, nil
}

// NewOrderEvent builds a CommerceEvent for the commerce.order.* family,
// encoding attrs into the envelope's Attributes field.
func NewOrderEvent(id, tenantID string, typ Type, source string, ts time.Time, attrs OrderAttributes) (CommerceEvent, error) {
	raw, err := json.Marshal(attrs)
	if err != nil {
		return CommerceEvent{}, err
	}
	return CommerceEvent{
		ID:         id,
		TenantID:   tenantID,
		Type:       typ,
		Source:     source,
		Timestamp:  ts,
		Version:    CurrentCommerceSchemaVersion,
		Attributes: raw,
	}, nil
}
