package postgres

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_run_event_storage.sql
var migration001 string

//go:embed migrations/002_transactional_outbox.sql
var migration002 string

func migrationParts(migration string) (string, string) {
	parts := strings.Split(migration, "-- migrate:down")
	return strings.TrimPrefix(parts[0], "-- migrate:up"), parts[1]
}

// MigrateUp installs the ordered schema. It is intentionally forward-only in
// production; MigrateDown exists to verify reversibility on clean test stores.
func MigrateUp(ctx context.Context, pool *pgxpool.Pool) error {
	up1, _ := migrationParts(migration001)
	up2, _ := migrationParts(migration002)
	_, err := pool.Exec(ctx, up1+"\n"+up2)
	return err
}

func MigrateDown(ctx context.Context, pool *pgxpool.Pool) error {
	_, down2 := migrationParts(migration002)
	_, down1 := migrationParts(migration001)
	_, err := pool.Exec(ctx, down2+"\n"+down1)
	return err
}

var errInvalidJSON = errors.New("postgres repository: invalid JSON document")
