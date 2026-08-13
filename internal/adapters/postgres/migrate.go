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

//go:embed migrations/003_run_supervisor.sql
var migration003 string

//go:embed migrations/004_input_delivery.sql
var migration004 string

//go:embed migrations/005_workflow_runtime.sql
var migration005 string

func migrationParts(migration string) (string, string) {
	parts := strings.Split(migration, "-- migrate:down")
	return strings.TrimPrefix(parts[0], "-- migrate:up"), parts[1]
}

// MigrateUp installs the ordered schema. It is intentionally forward-only in
// production; MigrateDown exists to verify reversibility on clean test stores.
func MigrateUp(ctx context.Context, pool *pgxpool.Pool) error {
	up1, _ := migrationParts(migration001)
	up2, _ := migrationParts(migration002)
	up3, _ := migrationParts(migration003)
	up4, _ := migrationParts(migration004)
	up5, _ := migrationParts(migration005)
	_, err := pool.Exec(ctx, up1+"\n"+up2+"\n"+up3+"\n"+up4+"\n"+up5)
	return err
}

func MigrateDown(ctx context.Context, pool *pgxpool.Pool) error {
	_, down5 := migrationParts(migration005)
	_, down4 := migrationParts(migration004)
	_, down3 := migrationParts(migration003)
	_, down2 := migrationParts(migration002)
	_, down1 := migrationParts(migration001)
	_, err := pool.Exec(ctx, down5+"\n"+down4+"\n"+down3+"\n"+down2+"\n"+down1)
	return err
}

var errInvalidJSON = errors.New("postgres repository: invalid JSON document")
