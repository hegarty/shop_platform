package event

import (
	"encoding/json"
	"time"

	"github.com/hegarty/shop_platform/money"
)

// Channel distinguishes a direct shop sale from a Shopify Collective sale.
// See shop_docs/docs/event-model.md for how this is derived from Shopify's
// order payload, and why ChannelUnknown exists.
type Channel string

const (
	ChannelShop       Channel = "shop"
	ChannelCollective Channel = "collective"
	// ChannelUnknown covers orders whose channel could not be confidently
	// classified from the source payload. Explicitly modeled rather than
	// defaulted to ChannelShop, so misclassification shows up as its own
	// queryable bucket instead of silently inflating direct sales.
	ChannelUnknown Channel = "unknown"
)

// OrderAttributes is the commerce.order.* event family's payload.
//
// Money fields are explicit and non-overlapping — there is deliberately no
// single ambiguous "sales" field. See shop_docs/docs/database-model.md for
// the exact definition of each:
//
//	GrossSales   sum of line item prices before discounts
//	Discounts    total discounts applied
//	Returns      total refunded amount
//	NetSales     GrossSales - Discounts - Returns
//	Shipping     shipping charged to the customer
//	Tax          tax collected
//	TotalSales   NetSales + Shipping + Tax (what the customer actually paid)
type OrderAttributes struct {
	ShopifyOrderID string     `json:"shopify_order_id"`
	OrderNumber    string     `json:"order_number"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`

	Currency string `json:"currency"`

	GrossSales money.Amount `json:"gross_sales"`
	Discounts  money.Amount `json:"discounts"`
	Returns    money.Amount `json:"returns"`
	NetSales   money.Amount `json:"net_sales"`
	Shipping   money.Amount `json:"shipping"`
	Tax        money.Amount `json:"tax"`
	TotalSales money.Amount `json:"total_sales"`

	Channel               Channel `json:"channel"`
	CollectivePartnerID   *string `json:"collective_partner_id,omitempty"`
	CollectivePartnerName *string `json:"collective_partner_name,omitempty"`

	FinancialStatus   string `json:"financial_status"`
	FulfillmentStatus string `json:"fulfillment_status"`

	// ClassificationVersion identifies which revision of the channel/partner
	// classification logic produced Channel/CollectivePartner*, so a later
	// fix can identify which historical events need reprocessing.
	ClassificationVersion int `json:"classification_version"`

	// RawMetadata preserves enough of the source Shopify payload to
	// investigate a classification error without re-fetching from Shopify.
	// Never exposed to consumers that don't need it (e.g. notifier).
	RawMetadata json.RawMessage `json:"raw_metadata,omitempty"`
}
