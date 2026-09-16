// Package money represents currency amounts as integer minor units (cents)
// so sales figures never accumulate floating-point rounding error across
// sums, refunds, and discounts.
package money

import (
	"fmt"
	"strconv"
	"strings"
)

// Amount is a quantity of currency in minor units (e.g. cents for USD).
// The currency itself is tracked alongside an Amount by the caller (see
// event.OrderAttributes.Currency) rather than embedded here, since almost
// every arithmetic operation on Amount is currency-agnostic.
type Amount int64

// FromDecimalString parses a decimal string such as "699.00" or "-12.5" (the
// format Shopify's Admin API uses for money fields) into minor units.
func FromDecimalString(s string) (Amount, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}

	whole, frac, hasFrac := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if !hasFrac {
		frac = "00"
	}
	switch len(frac) {
	case 0:
		frac = "00"
	case 1:
		frac += "0"
	case 2:
		// exact
	default:
		// truncate sub-cent precision rather than rounding silently wrong
		frac = frac[:2]
	}

	wholeVal, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("money: invalid decimal string %q: %w", s, err)
	}
	fracVal, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("money: invalid decimal string %q: %w", s, err)
	}

	total := wholeVal*100 + fracVal
	if neg {
		total = -total
	}
	return Amount(total), nil
}

// DecimalString renders the amount as a decimal string, e.g. "699.00".
func (a Amount) DecimalString() string {
	neg := a < 0
	v := int64(a)
	if neg {
		v = -v
	}
	s := fmt.Sprintf("%d.%02d", v/100, v%100)
	if neg {
		return "-" + s
	}
	return s
}

// MarshalJSON encodes the amount as a JSON string ("699.00"), matching the
// wire format of the systems this platform ingests from, and avoiding the
// float64-in-JSON precision trap entirely.
func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.DecimalString() + `"`), nil
}

// UnmarshalJSON accepts either a JSON string ("699.00") or a bare JSON
// number (699.0) for interoperability with producers that don't quote
// money fields.
func (a *Amount) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := FromDecimalString(s)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}

// Add returns the sum of two amounts.
func (a Amount) Add(b Amount) Amount { return a + b }

// Sub returns a minus b.
func (a Amount) Sub(b Amount) Amount { return a - b }

// Float64 converts to a float64 dollar (or other major-unit) value. Use only
// at presentation boundaries (formatting a human-readable summary) — never
// for further arithmetic, which should stay in Amount.
func (a Amount) Float64() float64 {
	return float64(a) / 100
}
