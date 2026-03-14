package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}

	wrapped := &DB{db}
	if err := wrapped.migrate(context.Background()); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return wrapped, nil
}

func (db *DB) migrate(ctx context.Context) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)
	`); err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		version := strings.TrimSuffix(entry.Name(), ".sql")

		var exists bool
		err := db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", version,
		).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}

		if _, err := db.ExecContext(ctx, string(data)); err != nil {
			return fmt.Errorf("apply migration %s: %w", version, err)
		}

		if _, err := db.ExecContext(ctx,
			"INSERT INTO schema_migrations(version) VALUES(?)", version,
		); err != nil {
			return err
		}
	}

	return nil
}
