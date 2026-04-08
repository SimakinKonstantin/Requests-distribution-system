package balancer

import (
	"context"
	"embed"
	"os"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func applyMigrations(ctx context.Context, db *DB) error {
	if os.Getenv("SKIP_MIGRATIONS") == "1" {
		return nil
	}

	// For demo: apply a single idempotent SQL file.
	b, err := migrationsFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}
	_, err = db.Pool.Exec(ctx, string(b))
	return err
}

