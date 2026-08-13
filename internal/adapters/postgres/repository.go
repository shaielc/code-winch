// Package postgres implements durable application ports without publishing
// events or importing another adapter.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

type Store struct{ pool *pgxpool.Pool }

type InputCommand struct {
	ID             domain.CommandID
	RunID          domain.RunID
	IdempotencyKey string
	Kind           string
	Payload        []byte
}

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// SaveResolvedConfiguration records canonical, fully resolved non-secret JSON
// separately from the existing RunRepository contract.
func (s *Store) SaveResolvedConfiguration(ctx context.Context, runID domain.RunID, configuration []byte) error {
	if !json.Valid(configuration) {
		return errInvalidJSON
	}
	tag, err := s.pool.Exec(ctx, `UPDATE runs SET resolved_configuration=$2 WHERE id=$1`, runID.String(), configuration)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

func conflict(err error, resource string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23514") {
		return fmt.Errorf("%w: resource=%s", application.ErrConflict, resource)
	}
	return err
}

func (s *Store) Save(ctx context.Context, record application.RunRecord, expected uint64) (uint64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	next := expected + 1
	if expected == 0 {
		_, err = tx.Exec(ctx, `INSERT INTO runs(id,version) VALUES($1,1)`, record.ID.String())
	} else {
		var found uint64
		err = tx.QueryRow(ctx, `UPDATE runs SET version=$2 WHERE id=$1 AND version=$3 RETURNING version`, record.ID.String(), next, expected).Scan(&found)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%w: run=%s expected_version=%d", application.ErrConflict, record.ID, expected)
		}
	}
	if err != nil {
		return 0, conflict(err, "run="+record.ID.String())
	}
	for i, a := range record.Attempts {
		_, err = tx.Exec(ctx, `INSERT INTO run_attempts(run_id,ordinal,id,previous_attempt_id,state) VALUES($1,$2,$3,NULLIF($4,'')::uuid,$5)
		ON CONFLICT(run_id,ordinal) DO UPDATE SET id=excluded.id,previous_attempt_id=excluded.previous_attempt_id,state=excluded.state`, record.ID.String(), i+1, a.ID.String(), zeroID(a.PreviousAttemptID), a.State)
		if err != nil {
			return 0, conflict(err, "run="+record.ID.String())
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, conflict(err, "run="+record.ID.String())
	}
	return next, nil
}

func zeroID(id domain.AttemptID) string {
	if id.IsZero() {
		return ""
	}
	return id.String()
}

func (s *Store) Get(ctx context.Context, id domain.RunID) (application.RunRecord, uint64, error) {
	var version uint64
	if err := s.pool.QueryRow(ctx, `SELECT version FROM runs WHERE id=$1`, id.String()).Scan(&version); errors.Is(err, pgx.ErrNoRows) {
		return application.RunRecord{}, 0, application.ErrNotFound
	} else if err != nil {
		return application.RunRecord{}, 0, err
	}
	rows, err := s.pool.Query(ctx, `SELECT id,COALESCE(previous_attempt_id::text,''),state FROM run_attempts WHERE run_id=$1 ORDER BY ordinal`, id.String())
	if err != nil {
		return application.RunRecord{}, 0, err
	}
	defer rows.Close()
	record := application.RunRecord{ID: id}
	for rows.Next() {
		var aid, prev, state string
		if err = rows.Scan(&aid, &prev, &state); err != nil {
			return application.RunRecord{}, 0, err
		}
		a, _ := domain.ParseAttemptID(aid)
		var p domain.AttemptID
		if prev != "" {
			p, _ = domain.ParseAttemptID(prev)
		}
		record.Attempts = append(record.Attempts, domain.Attempt{ID: a, PreviousAttemptID: p, State: domain.RunState(state)})
	}
	return record, version, rows.Err()
}

func (s *Store) Append(ctx context.Context, runID domain.RunID, expected uint64, values []application.UnsequencedEvent) ([]protocol.Event, error) {
	if len(values) == 0 {
		return []protocol.Event{}, nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var next uint64
	err = tx.QueryRow(ctx, `UPDATE runs SET last_sequence=last_sequence+$3 WHERE id=$1 AND last_sequence=$2 RETURNING last_sequence`, runID.String(), expected, len(values)).Scan(&next)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: run=%s expected_sequence=%d", application.ErrConflict, runID, expected)
	}
	if err != nil {
		return nil, err
	}
	out := make([]protocol.Event, len(values))
	for i, v := range values {
		if v.Sensitivity == protocol.SensitivitySecret {
			return nil, fmt.Errorf("postgres repository: secret event rejected: run=%s event=%s", runID, v.EventID)
		}
		seq := expected + uint64(i) + 1
		out[i] = application.NormalizeEvent(runID, seq, v, protocol.NormalizationPolicy{}).Event
		source, _ := json.Marshal(out[i].Source)
		ext, _ := json.Marshal(out[i].Extensions)
		_, err = tx.Exec(ctx, `INSERT INTO run_events(run_id,sequence,event_id,occurred_at,kind,schema_version,source,sensitivity,payload,extensions) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, runID.String(), seq, out[i].EventID, out[i].OccurredAt, out[i].Kind, out[i].SchemaVersion, source, out[i].Sensitivity, out[i].Payload, ext)
		if err != nil {
			return nil, conflict(err, "run="+runID.String())
		}
		envelope, marshalErr := json.Marshal(out[i])
		if marshalErr != nil {
			return nil, marshalErr
		}
		_, err = tx.Exec(ctx, `INSERT INTO outbox(id,topic,payload,run_id,sequence) VALUES($1,'run.events',$2,$3,$4)`, v.EventID.String(), envelope, runID.String(), seq)
		if err != nil {
			return nil, conflict(err, "run="+runID.String())
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, conflict(err, "run="+runID.String())
	}
	return out, nil
}

func (s *Store) ClaimOutbox(ctx context.Context, owner, token string, now, until domain.Timestamp, limit int) ([]application.OutboxRecord, error) {
	if limit <= 0 {
		return []application.OutboxRecord{}, nil
	}
	rows, err := s.pool.Query(ctx, `WITH candidates AS (
		SELECT o.id FROM outbox o WHERE o.completed_at IS NULL AND o.poisoned_at IS NULL AND o.available_at <= $3
		AND (o.lease_expires_at IS NULL OR o.lease_expires_at <= $3)
		AND NOT EXISTS (SELECT 1 FROM outbox prior WHERE prior.run_id=o.run_id AND prior.sequence<o.sequence AND prior.completed_at IS NULL AND prior.poisoned_at IS NULL)
		ORDER BY o.run_id, o.sequence FOR UPDATE OF o SKIP LOCKED LIMIT $4
	) UPDATE outbox o SET lease_owner=$1,lease_token=$2,lease_expires_at=$5,attempts=attempts+1
	FROM candidates c WHERE o.id=c.id RETURNING o.id,o.topic,o.payload,o.lease_token::text,o.attempts`, owner, token, now.Time(), limit, until.Time())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []application.OutboxRecord{}
	for rows.Next() {
		var id string
		var r application.OutboxRecord
		if err = rows.Scan(&id, &r.Message.Topic, &r.Message.Payload, &r.LeaseToken, &r.Attempts); err != nil {
			return nil, err
		}
		r.Message.ID, _ = domain.ParseCommandID(id)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) updateOutbox(ctx context.Context, id domain.CommandID, token string, query string, args ...any) error {
	params := []any{id.String(), token}
	params = append(params, args...)
	tag, err := s.pool.Exec(ctx, query, params...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: outbox=%s lease_token_stale", application.ErrConflict, id)
	}
	return nil
}

func (s *Store) CompleteOutbox(ctx context.Context, id domain.CommandID, token string, at domain.Timestamp) error {
	return s.updateOutbox(ctx, id, token, `UPDATE outbox SET completed_at=$3,lease_owner=NULL,lease_token=NULL,lease_expires_at=NULL WHERE id=$1 AND lease_token=$2 AND completed_at IS NULL`, at.Time())
}
func (s *Store) RetryOutbox(ctx context.Context, id domain.CommandID, token string, at domain.Timestamp, code string) error {
	return s.updateOutbox(ctx, id, token, `UPDATE outbox SET available_at=$3,last_error_code=$4,lease_owner=NULL,lease_token=NULL,lease_expires_at=NULL WHERE id=$1 AND lease_token=$2 AND completed_at IS NULL`, at.Time(), code)
}
func (s *Store) PoisonOutbox(ctx context.Context, id domain.CommandID, token string, at domain.Timestamp, code string) error {
	return s.updateOutbox(ctx, id, token, `UPDATE outbox SET poisoned_at=$3,last_error_code=$4,lease_owner=NULL,lease_token=NULL,lease_expires_at=NULL WHERE id=$1 AND lease_token=$2 AND completed_at IS NULL`, at.Time(), code)
}
func (s *Store) OutboxBacklog(ctx context.Context, now domain.Timestamp) (uint64, error) {
	var count uint64
	err := s.pool.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE completed_at IS NULL AND poisoned_at IS NULL AND available_at <= $1`, now.Time()).Scan(&count)
	return count, err
}

func (s *Store) Read(ctx context.Context, runID domain.RunID, after uint64, limit int) ([]protocol.Event, error) {
	if limit <= 0 {
		return []protocol.Event{}, nil
	}
	rows, err := s.pool.Query(ctx, `SELECT event_id,sequence,occurred_at,kind,schema_version,source,sensitivity,payload,extensions FROM run_events WHERE run_id=$1 AND sequence>$2 ORDER BY sequence LIMIT $3`, runID.String(), after, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []protocol.Event{}
	for rows.Next() {
		e := protocol.Event{RunID: runID.String()}
		if err = rows.Scan(&e.EventID, &e.Sequence, &e.OccurredAt, &e.Kind, &e.SchemaVersion, &e.Source, &e.Sensitivity, &e.Payload, &e.Extensions); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) PutCommand(ctx context.Context, command InputCommand) (bool, error) {
	if !json.Valid(command.Payload) {
		return false, errInvalidJSON
	}
	tag, err := s.pool.Exec(ctx, `INSERT INTO input_commands(id,run_id,idempotency_key,kind,payload) VALUES($1,$2,$3,$4,$5) ON CONFLICT(run_id,idempotency_key) DO NOTHING`, command.ID.String(), command.RunID.String(), command.IdempotencyKey, command.Kind, command.Payload)
	return tag.RowsAffected() == 1, err
}

func (s *Store) GetCommand(ctx context.Context, id domain.CommandID) (InputCommand, error) {
	command := InputCommand{ID: id}
	var runID string
	err := s.pool.QueryRow(ctx, `SELECT run_id,idempotency_key,kind,payload FROM input_commands WHERE id=$1`, id.String()).Scan(&runID, &command.IdempotencyKey, &command.Kind, &command.Payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return command, application.ErrNotFound
	}
	if err == nil {
		command.RunID, _ = domain.ParseRunID(runID)
	}
	return command, err
}
