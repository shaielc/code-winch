package postgres

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_run_event_storage.sql
var migration string

func migrationParts() (string, string) {
	parts := strings.Split(migration, "-- migrate:down")
	return strings.TrimPrefix(parts[0], "-- migrate:up"), parts[1]
}

// MigrateUp installs the ordered schema. It is intentionally forward-only in
// production; MigrateDown exists to verify reversibility on clean test stores.
func MigrateUp(ctx context.Context, pool *pgxpool.Pool) error {
	up, _ := migrationParts()
	_, err := pool.Exec(ctx, up)
	return err
}

func MigrateDown(ctx context.Context, pool *pgxpool.Pool) error {
	_, down := migrationParts()
	_, err := pool.Exec(ctx, down)
	return err
}

var errInvalidJSON = errors.New("postgres repository: invalid JSON document")
