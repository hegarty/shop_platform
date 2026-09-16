// Package db provides shared PostgreSQL helpers: connection pool setup and
// a minimal migration runner. Every table this platform uses is
// multi-tenant (tenant_id on every relevant row) — see
// shop_docs/docs/database-model.md.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds non-sensitive connection parameters. The password is never
// part of this struct — it's read from Secrets Manager at startup and
// passed to Connect separately, so it never ends up in a struct that might
// get logged or serialized by accident.
type Config struct {
	Host            string
	Port            int
	Database        string
	User            string
	MaxConns        int32
	MaxConnLifetime time.Duration
}

// Connect opens a pgx connection pool. password is kept out of Config so
// callers can log a Config value without redaction gymnastics.
func Connect(ctx context.Context, cfg Config, password string) (*pgxpool.Pool, error) {
	if cfg.MaxConns == 0 {
		cfg.MaxConns = 10
	}
	if cfg.MaxConnLifetime == 0 {
		cfg.MaxConnLifetime = 30 * time.Minute
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=require",
		cfg.User, password, cfg.Host, cfg.Port, cfg.Database,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("db: parse config: %w", err)
	}
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("db: create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	return pool, nil
}
