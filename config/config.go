// Package config loads non-sensitive configuration from environment
// variables and fails fast (at startup, not on first use) when something
// required is missing. Secrets are never read through this package — they
// come from AWS Secrets Manager; see shop_docs/docs/security.md.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Loader accumulates missing/invalid variables across multiple Require*
// calls so a service reports every configuration problem at once, instead
// of the developer fixing one env var, restarting, and hitting the next.
type Loader struct {
	errs []error
}

// NewLoader returns an empty Loader.
func NewLoader() *Loader {
	return &Loader{}
}

// String reads a required string env var.
func (l *Loader) String(name string) string {
	v, ok := os.LookupEnv(name)
	if !ok || v == "" {
		l.errs = append(l.errs, fmt.Errorf("missing required env var %s", name))
		return ""
	}
	return v
}

// StringDefault reads an optional string env var, returning def if unset.
func (l *Loader) StringDefault(name, def string) string {
	if v, ok := os.LookupEnv(name); ok && v != "" {
		return v
	}
	return def
}

// Int reads a required integer env var.
func (l *Loader) Int(name string) int {
	v, ok := os.LookupEnv(name)
	if !ok || v == "" {
		l.errs = append(l.errs, fmt.Errorf("missing required env var %s", name))
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("env var %s = %q is not a valid integer", name, v))
		return 0
	}
	return n
}

// IntDefault reads an optional integer env var, returning def if unset or
// invalid (invalid still records an error, so a typo doesn't silently fall
// back).
func (l *Loader) IntDefault(name string, def int) int {
	v, ok := os.LookupEnv(name)
	if !ok || v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("env var %s = %q is not a valid integer", name, v))
		return def
	}
	return n
}

// Bool reads an optional boolean env var (accepts the same formats as
// strconv.ParseBool), returning def if unset.
func (l *Loader) Bool(name string, def bool) bool {
	v, ok := os.LookupEnv(name)
	if !ok || v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("env var %s = %q is not a valid boolean", name, v))
		return def
	}
	return b
}

// Err returns a single combined error listing every missing/invalid
// variable encountered, or nil if everything loaded cleanly. Call this once
// after all Require*/Optional* calls, right before using any of the
// returned values.
func (l *Loader) Err() error {
	if len(l.errs) == 0 {
		return nil
	}
	msg := "config: invalid configuration:"
	for _, e := range l.errs {
		msg += "\n  - " + e.Error()
	}
	return fmt.Errorf("%s", msg)
}
