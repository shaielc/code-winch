//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	pgstore "github.com/shaielc/code-winch/internal/adapters/postgres"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

var (
	_ application.RunRepository     = (*pgstore.Store)(nil)
	_ application.EventStore        = (*pgstore.Store)(nil)
	_ application.OutboxStore       = (*pgstore.Store)(nil)
	_ application.InputCommandStore = (*pgstore.Store)(nil)
)

func TestInputAcceptanceIsAtomicConcurrentAndReplayable(t *testing.T) {
	pool, store := database(t)
	runID := id(t, domain.ParseRunID, 800)
	attemptID := id(t, domain.ParseAttemptID, 801)
	if _, err := store.Save(context.Background(), application.RunRecord{ID: runID, Attempts: []domain.Attempt{{ID: attemptID, State: domain.RunStateRunning}}}, 0); err != nil {
		t.Fatal(err)
	}
	seq := uint64(0)
	accept := application.InputAcceptance{Request: application.InputRequest{CommandID: id(t, domain.ParseCommandID, 802), RunID: runID, IdempotencyKey: "duplicate", ActorID: "actor", Kind: application.InputText, Payload: application.InputPayload{Text: "secret content"}, ExpectedState: domain.RunStateRunning, ExpectedSequence: &seq}, Capabilities: application.InputCapabilities{State: domain.RunStateRunning, Modes: map[application.InputKind]bool{application.InputText: true}}, DeliveryPayload: []byte(`{"commandId":"00000000-0000-0000-0000-000000000802","kind":"text","payload":{"text":"secret content"}}`)}
	type result struct {
		value application.InputResult
		err   error
	}
	results := make(chan result, 16)
	for range 16 {
		go func() { value, err := store.AcceptInput(context.Background(), accept); results <- result{value, err} }()
	}
	for range 16 {
		got := <-results
		if got.err != nil {
			t.Fatal(got.err)
		}
		if got.value.CommandID != accept.Request.CommandID {
			t.Fatalf("command=%s", got.value.CommandID)
		}
	}
	var commands, intents int
	_ = pool.QueryRow(context.Background(), `SELECT count(*) FROM input_commands`).Scan(&commands)
	_ = pool.QueryRow(context.Background(), `SELECT count(*) FROM outbox WHERE topic='run.input' AND completed_at IS NULL`).Scan(&intents)
	if commands != 1 || intents != 1 {
		t.Fatalf("commands=%d intents=%d", commands, intents)
	}
	// A daemon crash before delivery is represented by the still-uncompleted
	// intent; a new worker can claim it after restart.
	now, _ := domain.NewTimestamp(time.Now().UTC())
	until, _ := domain.NewTimestamp(time.Now().UTC().Add(time.Minute))
	claimed, err := store.ClaimOutbox(context.Background(), "restart-worker", "00000000-0000-0000-0000-000000000899", now, until, 10)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("restart claim=%d err=%v", len(claimed), err)
	}
}

func id[T any](t *testing.T, parse func(string) (T, error), n int) T {
	t.Helper()
	v, err := parse(fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestAppendCommitCreatesOutboxAndClaimsAreExclusive(t *testing.T) {
	pool, store := database(t)
	runID := id(t, domain.ParseRunID, 500)
	if _, err := store.Save(context.Background(), application.RunRecord{ID: runID}, 0); err != nil {
		t.Fatal(err)
	}
	occurredAt, _ := domain.NewTimestamp(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	values := []application.UnsequencedEvent{
		{EventID: id(t, domain.ParseEventID, 501), OccurredAt: occurredAt, Kind: "message", SchemaVersion: 1, Source: protocol.Source{Type: "test"}, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"n":1}`)},
		{EventID: id(t, domain.ParseEventID, 502), OccurredAt: occurredAt, Kind: "message", SchemaVersion: 1, Source: protocol.Source{Type: "test"}, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"n":2}`)},
	}
	if _, err := store.Append(context.Background(), runID, 0, values); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM outbox`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("outbox count=%d err=%v", count, err)
	}
	// Claims run on the wall clock, not the event clock: rows carry the
	// inserting transaction's own available_at, so a fixed date is never due.
	now, _ := domain.NewTimestamp(time.Now().UTC())
	until, _ := domain.NewTimestamp(time.Now().UTC().Add(time.Minute))
	type result struct {
		records []application.OutboxRecord
		err     error
	}
	results := make(chan result, 2)
	for n := 1; n <= 2; n++ {
		go func(n int) {
			records, err := store.ClaimOutbox(context.Background(), fmt.Sprintf("worker-%d", n), fmt.Sprintf("00000000-0000-0000-0000-%012d", 600+n), now, until, 1)
			results <- result{records, err}
		}(n)
	}
	total := 0
	for range 2 {
		r := <-results
		if r.err != nil {
			t.Fatal(r.err)
		}
		total += len(r.records)
	}
	if total != 1 {
		t.Fatalf("concurrent workers claimed %d ordered records, want 1", total)
	}
}

func TestRollbackLeavesNoEventOrPublishIntent(t *testing.T) {
	pool, store := database(t)
	runID := id(t, domain.ParseRunID, 700)
	if _, err := store.Save(context.Background(), application.RunRecord{ID: runID}, 0); err != nil {
		t.Fatal(err)
	}
	now, _ := domain.NewTimestamp(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	// The second event duplicates the first ID, forcing the entire append transaction to roll back.
	event := application.UnsequencedEvent{EventID: id(t, domain.ParseEventID, 701), OccurredAt: now, Kind: "message", SchemaVersion: 1, Source: protocol.Source{Type: "test"}, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{}`)}
	if _, err := store.Append(context.Background(), runID, 0, []application.UnsequencedEvent{event, event}); err == nil {
		t.Fatal("append succeeded")
	}
	var events, outbox int
	_ = pool.QueryRow(context.Background(), `SELECT count(*) FROM run_events`).Scan(&events)
	_ = pool.QueryRow(context.Background(), `SELECT count(*) FROM outbox`).Scan(&outbox)
	if events != 0 || outbox != 0 {
		t.Fatalf("rollback events=%d outbox=%d", events, outbox)
	}
}

func database(t *testing.T) (*pgxpool.Pool, *pgstore.Store) {
	t.Helper()
	url := os.Getenv("PG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("PG_TEST_DATABASE_URL is required")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	_, _ = pool.Exec(context.Background(), `DROP SCHEMA public CASCADE; CREATE SCHEMA public`)
	if _, err = pgstore.MigrateUp(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	return pool, pgstore.New(pool)
}

func TestMigrationRoundTripOnCleanDatabase(t *testing.T) {
	pool, _ := database(t)
	if err := pgstore.MigrateDown(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pgstore.MigrateUp(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
}

// The daemon migrates on every start, so a second start against an already
// migrated database must change nothing rather than fail on an existing table.
func TestMigrateUpOnMigratedDatabaseIsANoOp(t *testing.T) {
	pool, _ := database(t)
	result, err := pgstore.MigrateUp(context.Background(), pool)
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if result.Applied != 0 {
		t.Fatalf("second migrate applied %d migrations, want 0", result.Applied)
	}
	var applied int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatal(err)
	}
	// Counted from the directory rather than hardcoded: a migration nobody
	// applied is exactly the defect this test exists to catch.
	files, err := os.ReadDir("migrations")
	if err != nil {
		t.Fatal(err)
	}
	if applied != len(files) {
		t.Fatalf("ledger records %d migrations for %d files", applied, len(files))
	}
}

func TestRunHistoryConfigurationAndCommands(t *testing.T) {
	pool, store := database(t)
	runID := id(t, domain.ParseRunID, 1)
	aid := id(t, domain.ParseAttemptID, 2)
	record := application.RunRecord{ID: runID, Attempts: []domain.Attempt{{ID: aid, State: domain.RunStateFailed}}}
	version, err := store.Save(context.Background(), record, 0)
	if err != nil || version != 1 {
		t.Fatalf("save: version=%d err=%v", version, err)
	}
	got, gotVersion, err := store.Get(context.Background(), runID)
	if err != nil || gotVersion != 1 || len(got.Attempts) != 1 {
		t.Fatalf("get: %#v %d %v", got, gotVersion, err)
	}
	if err = store.SaveResolvedConfiguration(context.Background(), runID, []byte(`{"profile":"safe","credentialRef":"vault:item"}`)); err != nil {
		t.Fatal(err)
	}
	record.Attempts[0].State = domain.RunStateRunning
	if _, err = store.Save(context.Background(), record, 1); !errors.Is(err, application.ErrConflict) {
		t.Fatalf("terminal mutation: %v", err)
	}

	if _, err = pool.Exec(context.Background(), `UPDATE run_attempts SET state='running' WHERE run_id=$1`, runID.String()); err == nil {
		t.Fatal("database allowed terminal history mutation")
	}
	command := pgstore.InputCommand{ID: id(t, domain.ParseCommandID, 3), RunID: runID, IdempotencyKey: "client-1", Kind: "input", Payload: []byte(`{"text":"hello"}`)}
	created, err := store.PutCommand(context.Background(), command)
	if err != nil || !created {
		t.Fatalf("command create: %v %v", created, err)
	}
	created, err = store.PutCommand(context.Background(), command)
	if err != nil || created {
		t.Fatalf("command replay: %v %v", created, err)
	}
}

func TestConcurrentAppendIsGapFreeAndSecretCanaryAbsent(t *testing.T) {
	pool, store := database(t)
	runID := id(t, domain.ParseRunID, 10)
	if _, err := store.Save(context.Background(), application.RunRecord{ID: runID}, 0); err != nil {
		t.Fatal(err)
	}
	now, _ := domain.NewTimestamp(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	const writers = 32
	var wg sync.WaitGroup
	results := make(chan error, writers)
	for n := 1; n <= writers; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for {
				current, err := store.Read(context.Background(), runID, 0, writers+1)
				if err != nil {
					results <- err
					return
				}
				_, err = store.Append(context.Background(), runID, uint64(len(current)), []application.UnsequencedEvent{{EventID: id(t, domain.ParseEventID, 100+n), OccurredAt: now, Kind: "message", SchemaVersion: 1, Source: protocol.Source{Type: "test"}, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"text":"safe"}`)}})
				if errors.Is(err, application.ErrConflict) {
					continue
				}
				results <- err
				return
			}
		}(n)
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else {
			t.Fatal(err)
		}
	}
	if success != writers {
		t.Fatalf("success=%d", success)
	}
	events, err := store.Read(context.Background(), runID, 0, 100)
	if err != nil || len(events) != writers {
		t.Fatalf("events=%#v err=%v", events, err)
	}
	for i, event := range events {
		if event.Sequence != uint64(i+1) {
			t.Fatalf("sequence[%d]=%d", i, event.Sequence)
		}
	}
	canary := "SECRET_CANARY_DO_NOT_STORE"
	_, err = store.Append(context.Background(), runID, writers, []application.UnsequencedEvent{{EventID: id(t, domain.ParseEventID, 999), OccurredAt: now, Kind: "message", SchemaVersion: 1, Source: protocol.Source{Type: "test"}, Sensitivity: protocol.SensitivitySecret, Payload: []byte(fmt.Sprintf(`{"value":%q}`, canary))}})
	if err == nil {
		t.Fatal("secret event accepted")
	}
	var dump string
	if err = pool.QueryRow(context.Background(), `SELECT COALESCE(string_agg(payload::text,''),'') FROM run_events`).Scan(&dump); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(dump, canary) {
		t.Fatal("secret canary persisted")
	}
}
