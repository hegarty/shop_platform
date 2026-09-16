package money

import "testing"

func TestFromDecimalString(t *testing.T) {
	cases := []struct {
		in   string
		want Amount
	}{
		{"699.00", 69900},
		{"699", 69900},
		{"0.05", 5},
		{"-12.5", -1250},
		{"", 0},
		{"10.999", 1099}, // sub-cent precision truncated, not rounded up
	}
	for _, c := range cases {
		got, err := FromDecimalString(c.in)
		if err != nil {
			t.Fatalf("FromDecimalString(%q) error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("FromDecimalString(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestFromDecimalString_Invalid(t *testing.T) {
	if _, err := FromDecimalString("not-a-number"); err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestDecimalStringRoundTrip(t *testing.T) {
	cases := []string{"699.00", "0.05", "-12.50", "0.00"}
	for _, c := range cases {
		amt, err := FromDecimalString(c)
		if err != nil {
			t.Fatalf("FromDecimalString(%q): %v", c, err)
		}
		if got := amt.DecimalString(); got != c {
			t.Errorf("round trip %q -> %d -> %q, want %q", c, amt, got, c)
		}
	}
}

func TestJSONRoundTrip(t *testing.T) {
	amt, _ := FromDecimalString("1282.50")
	b, err := amt.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(b) != `"1282.50"` {
		t.Errorf("MarshalJSON = %s, want \"1282.50\"", b)
	}

	var out Amount
	if err := out.UnmarshalJSON(b); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if out != amt {
		t.Errorf("UnmarshalJSON round trip = %d, want %d", out, amt)
	}
}

func TestUnmarshalJSON_BareNumber(t *testing.T) {
	var a Amount
	if err := a.UnmarshalJSON([]byte("699.5")); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if a != 69950 {
		t.Errorf("got %d, want 69950", a)
	}
}

func TestAddSub(t *testing.T) {
	a, _ := FromDecimalString("100.00")
	b, _ := FromDecimalString("30.00")
	if got := a.Add(b); got != 13000 {
		t.Errorf("Add = %d, want 13000", got)
	}
	if got := a.Sub(b); got != 7000 {
		t.Errorf("Sub = %d, want 7000", got)
	}
}
