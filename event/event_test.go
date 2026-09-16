package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hegarty/shop_platform/money"
)

func TestRawEnvelope_Validate(t *testing.T) {
	valid := RawEnvelope{
		EventID:   "evt-1",
		TenantID:  "devmoto",
		Source:    "shopify",
		EventType: "orders/create",
		Payload:   json.RawMessage(`{}`),
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid envelope, got error: %v", err)
	}

	cases := []struct {
		name    string
		mutate  func(e RawEnvelope) RawEnvelope
		wantErr error
	}{
		{"missing event id", func(e RawEnvelope) RawEnvelope { e.EventID = ""; return e }, ErrMissingEventID},
		{"missing tenant", func(e RawEnvelope) RawEnvelope { e.TenantID = ""; return e }, ErrMissingTenantID},
		{"missing source", func(e RawEnvelope) RawEnvelope { e.Source = ""; return e }, ErrMissingSource},
		{"missing event type", func(e RawEnvelope) RawEnvelope { e.EventType = ""; return e }, ErrMissingEventType},
		{"missing payload", func(e RawEnvelope) RawEnvelope { e.Payload = nil; return e }, ErrMissingPayload},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.mutate(valid).Validate(); err != c.wantErr {
				t.Errorf("got %v, want %v", err, c.wantErr)
			}
		})
	}
}

func TestNewOrderEvent_RoundTrip(t *testing.T) {
	gross, _ := money.FromDecimalString("699.00")
	net, _ := money.FromDecimalString("649.00")
	partner := "ABC Bikes"

	attrs := OrderAttributes{
		ShopifyOrderID:        "gid://shopify/Order/123",
		OrderNumber:           "#1042",
		CreatedAt:             time.Date(2026, 9, 11, 15, 31, 12, 0, time.UTC),
		Currency:              "USD",
		GrossSales:            gross,
		NetSales:              net,
		Channel:               ChannelCollective,
		CollectivePartnerName: &partner,
		FinancialStatus:       "paid",
		FulfillmentStatus:     "unfulfilled",
		ClassificationVersion: 1,
	}

	evt, err := NewOrderEvent("evt-1", "devmoto", TypeOrderCreated, "shopify", attrs.CreatedAt, attrs)
	if err != nil {
		t.Fatalf("NewOrderEvent: %v", err)
	}
	if err := evt.Validate(); err != nil {
		t.Fatalf("expected valid event, got error: %v", err)
	}
	if evt.Version != CurrentCommerceSchemaVersion {
		t.Errorf("Version = %d, want %d", evt.Version, CurrentCommerceSchemaVersion)
	}

	// Round-trip through JSON, as it would cross Redpanda.
	wire, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded CommerceEvent
	if err := json.Unmarshal(wire, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	got, err := decoded.DecodeOrderAttributes()
	if err != nil {
		t.Fatalf("DecodeOrderAttributes: %v", err)
	}
	if got.Channel != ChannelCollective {
		t.Errorf("Channel = %q, want %q", got.Channel, ChannelCollective)
	}
	if got.CollectivePartnerName == nil || *got.CollectivePartnerName != partner {
		t.Errorf("CollectivePartnerName = %v, want %q", got.CollectivePartnerName, partner)
	}
	if got.GrossSales != gross {
		t.Errorf("GrossSales = %d, want %d", got.GrossSales, gross)
	}
}

func TestCommerceEvent_Validate(t *testing.T) {
	valid := CommerceEvent{
		ID:         "evt-1",
		TenantID:   "devmoto",
		Type:       TypeOrderCreated,
		Attributes: json.RawMessage(`{}`),
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}

	missing := valid
	missing.Attributes = nil
	if err := missing.Validate(); err != ErrMissingAttributes {
		t.Errorf("got %v, want ErrMissingAttributes", err)
	}
}
