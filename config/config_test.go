package config

import (
	"strings"
	"testing"
)

func TestLoader_AccumulatesAllErrors(t *testing.T) {
	l := NewLoader()
	l.String("SHOP_TEST_MISSING_STRING")
	l.Int("SHOP_TEST_MISSING_INT")

	err := l.Err()
	if err == nil {
		t.Fatal("expected an error")
	}
	got := err.Error()
	if !strings.Contains(got, "SHOP_TEST_MISSING_STRING") || !strings.Contains(got, "SHOP_TEST_MISSING_INT") {
		t.Errorf("expected both missing vars mentioned, got: %s", got)
	}
}

func TestLoader_ValidValues(t *testing.T) {
	t.Setenv("SHOP_TEST_HOST", "db.internal")
	t.Setenv("SHOP_TEST_PORT", "5432")

	l := NewLoader()
	host := l.String("SHOP_TEST_HOST")
	port := l.Int("SHOP_TEST_PORT")

	if err := l.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host != "db.internal" {
		t.Errorf("host = %q", host)
	}
	if port != 5432 {
		t.Errorf("port = %d", port)
	}
}

func TestLoader_InvalidIntRecordsError(t *testing.T) {
	t.Setenv("SHOP_TEST_BAD_INT", "not-a-number")

	l := NewLoader()
	l.Int("SHOP_TEST_BAD_INT")

	if err := l.Err(); err == nil {
		t.Fatal("expected error for invalid integer")
	}
}

func TestLoader_Defaults(t *testing.T) {
	l := NewLoader()
	if got := l.StringDefault("SHOP_TEST_UNSET", "fallback"); got != "fallback" {
		t.Errorf("StringDefault = %q, want fallback", got)
	}
	if got := l.IntDefault("SHOP_TEST_UNSET_INT", 42); got != 42 {
		t.Errorf("IntDefault = %d, want 42", got)
	}
	if got := l.Bool("SHOP_TEST_UNSET_BOOL", true); got != true {
		t.Errorf("Bool = %v, want true", got)
	}
	if err := l.Err(); err != nil {
		t.Fatalf("unexpected error from defaults-only loader: %v", err)
	}
}
