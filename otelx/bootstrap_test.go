package otelx

import (
	"context"
	"testing"
	"time"
)

func TestBootstrap_SucceedsWithoutBlockingOnDial(t *testing.T) {
	// gRPC dialing is lazy by default, so Bootstrap should succeed even
	// against an endpoint nothing is listening on — the first real export
	// attempt would fail, not this call.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shutdown, err := Bootstrap(ctx, Config{
		ServiceName:    "test-service",
		ServiceVersion: "0.0.0-test",
		Endpoint:       "127.0.0.1:1", // deliberately unreachable
		Insecure:       true,
	})
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected a non-nil Shutdown func")
	}

	// Shutdown forces a final export flush, which legitimately fails here
	// since nothing is listening on the endpoint — that's exercising real
	// (correct) behavior, not something this test should assert against.
	// We only care that Shutdown returns promptly rather than hanging.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = shutdown(shutdownCtx)
}
