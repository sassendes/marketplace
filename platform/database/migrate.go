package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
)

const migrationLockKey = 727274001

type Migration struct {
	Version  string
	SQL      string
	Checksum string
}

func LoadMigrations(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	var out []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		b, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(b)
		out = append(out, Migration{
			Version:  strings.TrimSuffix(e.Name(), ".sql"),
			SQL:      string(b),
			Checksum: hex.EncodeToString(sum[:]),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

func Migrate(ctx context.Context, db *sql.DB, migrations []Migration) (int, error) {
	count := 0
	for _, m := range migrations {
		applied, err := applyMigration(ctx, db, m)
		if err != nil {
			return count, err
		}
		if applied {
			slog.Info("applied migration", "version", m.Version)
			count++
		}
	}
	return count, nil
}

func applyMigration(ctx context.Context, db *sql.DB, m Migration) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, migrationLockKey); err != nil {
		return false, fmt.Errorf("acquire migration lock: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			checksum   TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return false, fmt.Errorf("create schema_migrations: %w", err)
	}

	var sum string
	err = tx.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE version = $1`, m.Version).Scan(&sum)
	switch {
	case err == nil:
		if sum != m.Checksum {
			return false, fmt.Errorf("migration %s was modified after it was applied (checksum mismatch)", m.Version)
		}
		return false, tx.Commit()
	case err != sql.ErrNoRows:
		return false, err
	}

	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		return false, fmt.Errorf("apply migration %s: %w", m.Version, err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)`, m.Version, m.Checksum); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func PendingMigrations(ctx context.Context, db *sql.DB, migrations []Migration) ([]string, error) {
	var exists bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations')`).Scan(&exists); err != nil {
		return nil, err
	}
	applied := map[string]bool{}
	if exists {
		rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var v string
			if err := rows.Scan(&v); err != nil {
				return nil, err
			}
			applied[v] = true
		}
	}
	var pending []string
	for _, m := range migrations {
		if !applied[m.Version] {
			pending = append(pending, m.Version)
		}
	}
	return pending, nil
}
