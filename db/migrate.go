package db

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration is one numbered, named schema change. Deliberately a small,
// hand-rolled runner rather than a full migration framework dependency —
// this platform only ever targets Postgres, so the extra abstraction a
// multi-database framework provides isn't buying anything.
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

var migrationFileRe = regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

// LoadMigrations reads migration files from fsys. Expected naming:
// 0001_create_tenants.up.sql / 0001_create_tenants.down.sql. Every "up"
// file must have a matching "down" file — a migration with no rollback
// path is a foot-gun, so this is a load-time error, not a runtime surprise.
func LoadMigrations(fsys fs.FS) ([]Migration, error) {
	byVersion := map[int]*Migration{}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("db: read migrations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		m := migrationFileRe.FindStringSubmatch(entry.Name())
		if m == nil {
			continue
		}
		version, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("db: invalid migration version in %q: %w", entry.Name(), err)
		}
		name, direction := m[2], m[3]

		content, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("db: read %q: %w", entry.Name(), err)
		}

		mig, ok := byVersion[version]
		if !ok {
			mig = &Migration{Version: version, Name: name}
			byVersion[version] = mig
		}
		if direction == "up" {
			mig.Up = string(content)
		} else {
			mig.Down = string(content)
		}
	}

	migrations := make([]Migration, 0, len(byVersion))
	for _, m := range byVersion {
		if m.Up == "" {
			return nil, fmt.Errorf("db: migration %04d (%s) has a down file but no up file", m.Version, m.Name)
		}
		if m.Down == "" {
			return nil, fmt.Errorf("db: migration %04d (%s) has no down file — every migration must be reversible", m.Version, m.Name)
		}
		migrations = append(migrations, *m)
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })

	return migrations, nil
}

const createSchemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version     INTEGER PRIMARY KEY,
	name        TEXT NOT NULL,
	applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);`

// Migrate applies every pending migration, in order, each in its own
// transaction. Safe to call repeatedly — already-applied versions are
// skipped.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrations []Migration) error {
	if _, err := pool.Exec(ctx, createSchemaMigrationsTable); err != nil {
		return fmt.Errorf("db: ensure schema_migrations table: %w", err)
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("db: begin tx for migration %04d: %w", m.Version, err)
		}

		if _, err := tx.Exec(ctx, m.Up); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("db: apply migration %04d (%s): %w", m.Version, m.Name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.Version, m.Name); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("db: record migration %04d: %w", m.Version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("db: commit migration %04d: %w", m.Version, err)
		}
	}

	return nil
}

// Rollback reverts the most recently applied `steps` migrations, most
// recent first, each in its own transaction.
func Rollback(ctx context.Context, pool *pgxpool.Pool, migrations []Migration, steps int) error {
	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return err
	}

	byVersion := make(map[int]Migration, len(migrations))
	for _, m := range migrations {
		byVersion[m.Version] = m
	}

	var appliedList []int
	for v := range applied {
		appliedList = append(appliedList, v)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(appliedList)))

	for i, v := range appliedList {
		if i >= steps {
			break
		}
		m, ok := byVersion[v]
		if !ok {
			return fmt.Errorf("db: applied migration %04d has no matching loaded migration to roll back", v)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("db: begin tx for rollback %04d: %w", v, err)
		}
		if _, err := tx.Exec(ctx, m.Down); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("db: rollback migration %04d (%s): %w", v, m.Name, err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, v); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("db: unrecord migration %04d: %w", v, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("db: commit rollback %04d: %w", v, err)
		}
	}

	return nil
}

func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[int]bool, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("db: query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("db: scan applied version: %w", err)
		}
		applied[v] = true
	}
	return applied, rows.Err()
}
