// Package tenant defines the shared Tenant shape every shop_* service
// reads from the tenants table. Deliberately minimal — this platform is
// multi-tenant from day one (every event and query is tenant-scoped), but
// does not yet implement a full user/identity system on top of it. See
// ADR-007.
package tenant

// Tenant identifies one business using the platform (e.g. "devmoto").
type Tenant struct {
	ID         string
	Name       string
	Timezone   string // IANA name, e.g. "America/New_York" — see period.Resolve
	ShopDomain string // primary Shopify shop domain, e.g. "devmoto.myshopify.com"
}
