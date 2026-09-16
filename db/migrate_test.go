package db

import (
	"testing"
	"testing/fstest"
)

func TestLoadMigrations_OrdersByVersion(t *testing.T) {
	fsys := fstest.MapFS{
		"0002_add_index.up.sql":        {Data: []byte("CREATE INDEX x ON y (z);")},
		"0002_add_index.down.sql":      {Data: []byte("DROP INDEX x;")},
		"0001_create_tenants.up.sql":   {Data: []byte("CREATE TABLE tenants ();")},
		"0001_create_tenants.down.sql": {Data: []byte("DROP TABLE tenants;")},
	}

	migrations, err := LoadMigrations(fsys)
	if err != nil {
		t.Fatalf("LoadMigrations: %v", err)
	}
	if len(migrations) != 2 {
		t.Fatalf("got %d migrations, want 2", len(migrations))
	}
	if migrations[0].Version != 1 || migrations[1].Version != 2 {
		t.Errorf("migrations not sorted by version: %+v", migrations)
	}
	if migrations[0].Name != "create_tenants" {
		t.Errorf("Name = %q, want %q", migrations[0].Name, "create_tenants")
	}
}

func TestLoadMigrations_MissingDownFile(t *testing.T) {
	fsys := fstest.MapFS{
		"0001_create_tenants.up.sql": {Data: []byte("CREATE TABLE tenants ();")},
	}
	if _, err := LoadMigrations(fsys); err == nil {
		t.Fatal("expected error for migration missing a down file")
	}
}

func TestLoadMigrations_IgnoresUnrelatedFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"0001_create_tenants.up.sql":   {Data: []byte("CREATE TABLE tenants ();")},
		"0001_create_tenants.down.sql": {Data: []byte("DROP TABLE tenants;")},
		"README.md":                    {Data: []byte("not a migration")},
	}
	migrations, err := LoadMigrations(fsys)
	if err != nil {
		t.Fatalf("LoadMigrations: %v", err)
	}
	if len(migrations) != 1 {
		t.Fatalf("got %d migrations, want 1", len(migrations))
	}
}
