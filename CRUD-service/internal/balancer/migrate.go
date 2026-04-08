package main

import (
	"context"
	"os"
	"path/filepath"
)

func applyMigrations(ctx context.Context, db *DB) error {
	if os.Getenv("SKIP_MIGRATIONS") == "1" {
		return nil
	}

	// For demo: apply a single idempotent SQL file.
	path := filepath.Join("migrations", "001_init.sql")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = db.Pool.Exec(ctx, string(b))
	return err
}

