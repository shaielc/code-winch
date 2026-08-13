package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

func (s *Store) PutWorkflowDefinition(ctx context.Context, record application.WorkflowDefinitionRecord) error {
	if !json.Valid(record.Document) {
		return errInvalidJSON
	}
	tag, err := s.pool.Exec(ctx, `INSERT INTO workflow_definitions(definition_id,version,document,created_at)
		VALUES($1,$2,$3,$4) ON CONFLICT(definition_id,version) DO NOTHING`, record.DefinitionID, record.Version, record.Document, record.CreatedAt.Time())
	if err != nil {
		return conflict(err, "workflow_definition="+record.DefinitionID+" version="+record.Version)
	}
	if tag.RowsAffected() != 0 {
		return nil
	}
	var same bool
	if err = s.pool.QueryRow(ctx, `SELECT document=$3::jsonb FROM workflow_definitions WHERE definition_id=$1 AND version=$2`, record.DefinitionID, record.Version, record.Document).Scan(&same); err != nil {
		return err
	}
	if !same {
		return fmt.Errorf("%w: workflow_definition=%s version=%s immutable", application.ErrConflict, record.DefinitionID, record.Version)
	}
	return nil
}

func (s *Store) CreateWorkflowInstance(ctx context.Context, instance application.WorkflowInstanceRecord, attempts []application.WorkflowStepAttempt) error {
	if !json.Valid(instance.Inputs) {
		return errInvalidJSON
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, `INSERT INTO workflow_instances(id,definition_id,definition_version,status,inputs,harness_profile,sandbox_profile,parent_instance_id,parent_step_id,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($8,'')::uuid,NULLIF($9,''),$10)`, instance.ID.String(), instance.DefinitionID, instance.DefinitionVersion, instance.Status, instance.Inputs, instance.HarnessProfile, instance.SandboxProfile, workflowID(instance.ParentInstanceID), instance.ParentStepID, instance.CreatedAt.Time())
	if err != nil {
		return conflict(err, "workflow="+instance.ID.String())
	}
	for _, attempt := range attempts {
		if attempt.InstanceID != instance.ID {
			return fmt.Errorf("%w: workflow=%s attempt_instance_mismatch", application.ErrConflict, instance.ID)
		}
		if err = insertWorkflowAttempt(ctx, tx, attempt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func workflowID(id domain.WorkflowID) string {
	if id.IsZero() {
		return ""
	}
	return id.String()
}

// queryExecer names only the pgx transaction operation used here, keeping the
// insertion helper usable inside the surrounding atomic transaction.
type queryExecer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func insertWorkflowAttempt(ctx context.Context, tx queryExecer, a application.WorkflowStepAttempt) error {
	var started, finished any
	if a.StartedAt != nil {
		started = a.StartedAt.Time()
	}
	if a.FinishedAt != nil {
		finished = a.FinishedAt.Time()
	}
	var output any
	if len(a.Output) != 0 {
		if !json.Valid(a.Output) {
			return errInvalidJSON
		}
		output = a.Output
	}
	_, err := tx.Exec(ctx, `INSERT INTO workflow_step_attempts(id,workflow_instance_id,step_id,attempt,state,available_at,started_at,finished_at,output,error_code)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, a.ID.String(), a.InstanceID.String(), a.StepID, a.Attempt, a.State, a.AvailableAt.Time(), started, finished, output, emptyNull(a.ErrorCode))
	return conflict(err, "workflow="+a.InstanceID.String()+" step="+a.StepID)
}

func emptyNull(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (s *Store) AppendWorkflowAttempt(ctx context.Context, attempt application.WorkflowStepAttempt) error {
	return insertWorkflowAttempt(ctx, s.pool, attempt)
}

func (s *Store) GetWorkflowInstance(ctx context.Context, id domain.WorkflowID) (application.WorkflowInstanceRecord, []application.WorkflowStepAttempt, error) {
	r := application.WorkflowInstanceRecord{ID: id}
	var parent string
	var created time.Time
	err := s.pool.QueryRow(ctx, `SELECT definition_id,definition_version,status,inputs,harness_profile,sandbox_profile,COALESCE(parent_instance_id::text,''),COALESCE(parent_step_id,''),created_at FROM workflow_instances WHERE id=$1`, id.String()).Scan(&r.DefinitionID, &r.DefinitionVersion, &r.Status, &r.Inputs, &r.HarnessProfile, &r.SandboxProfile, &parent, &r.ParentStepID, &created)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, nil, application.ErrNotFound
	}
	if err != nil {
		return r, nil, err
	}
	r.CreatedAt = mustTimestamp(created)
	if parent != "" {
		r.ParentInstanceID, _ = domain.ParseWorkflowID(parent)
	}
	rows, err := s.pool.Query(ctx, `SELECT id,step_id,attempt,state,available_at,started_at,finished_at,COALESCE(output,'null'::jsonb),COALESCE(error_code,''),COALESCE(lease_owner,''),COALESCE(lease_token::text,''),lease_epoch,lease_expires_at FROM workflow_step_attempts WHERE workflow_instance_id=$1 ORDER BY step_id,attempt`, id.String())
	if err != nil {
		return r, nil, err
	}
	defer rows.Close()
	out := []application.WorkflowStepAttempt{}
	for rows.Next() {
		a := application.WorkflowStepAttempt{InstanceID: id}
		var aid string
		var available, started, finished, expires pgtype.Timestamptz
		if err = rows.Scan(&aid, &a.StepID, &a.Attempt, &a.State, &available, &started, &finished, &a.Output, &a.ErrorCode, &a.LeaseOwner, &a.LeaseToken, &a.LeaseEpoch, &expires); err != nil {
			return r, nil, err
		}
		a.ID, _ = domain.ParseAttemptID(aid)
		a.AvailableAt = mustTimestamp(available.Time)
		a.StartedAt = optionalTimestamp(started)
		a.FinishedAt = optionalTimestamp(finished)
		a.LeaseExpiresAt = optionalTimestamp(expires)
		out = append(out, a)
	}
	return r, out, rows.Err()
}

func mustTimestamp(value time.Time) domain.Timestamp { v, _ := domain.NewTimestamp(value); return v }
func optionalTimestamp(value pgtype.Timestamptz) *domain.Timestamp {
	if !value.Valid {
		return nil
	}
	v := mustTimestamp(value.Time)
	return &v
}

func (s *Store) ClaimReadyWorkflowSteps(ctx context.Context, owner, token string, now, until domain.Timestamp, limit int) ([]application.WorkflowStepAttempt, error) {
	if limit <= 0 {
		return []application.WorkflowStepAttempt{}, nil
	}
	rows, err := s.pool.Query(ctx, `WITH candidates AS (
		SELECT a.id FROM workflow_step_attempts a JOIN workflow_instances i ON i.id=a.workflow_instance_id
		WHERE i.status IN ('pending','running') AND a.available_at <= $3
		AND (a.state='ready' OR (a.state='running' AND a.lease_expires_at <= $3))
		ORDER BY a.available_at,a.workflow_instance_id,a.step_id FOR UPDATE OF a SKIP LOCKED LIMIT $5
	) UPDATE workflow_step_attempts a SET state='running',started_at=COALESCE(started_at,$3),lease_owner=$1,lease_token=$2,lease_epoch=lease_epoch+1,lease_expires_at=$4
	FROM candidates c WHERE a.id=c.id RETURNING a.id,a.workflow_instance_id,a.step_id,a.attempt,a.state,a.available_at,a.started_at,a.lease_owner,a.lease_token::text,a.lease_epoch,a.lease_expires_at`, owner, token, now.Time(), until.Time(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []application.WorkflowStepAttempt{}
	for rows.Next() {
		var a application.WorkflowStepAttempt
		var aid, wid string
		var available, started, expires pgtype.Timestamptz
		if err = rows.Scan(&aid, &wid, &a.StepID, &a.Attempt, &a.State, &available, &started, &a.LeaseOwner, &a.LeaseToken, &a.LeaseEpoch, &expires); err != nil {
			return nil, err
		}
		a.ID, _ = domain.ParseAttemptID(aid)
		a.InstanceID, _ = domain.ParseWorkflowID(wid)
		a.AvailableAt = mustTimestamp(available.Time)
		a.StartedAt = optionalTimestamp(started)
		a.LeaseExpiresAt = optionalTimestamp(expires)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) CompleteWorkflowStep(ctx context.Context, lease application.WorkflowStepAttempt, state string, output json.RawMessage, errorCode string, at domain.Timestamp, messages []application.WorkflowOutboxMessage) error {
	if state != "completed" && state != "failed" && state != "cancelled" {
		return fmt.Errorf("%w: workflow=%s step=%s invalid_terminal_state", application.ErrConflict, lease.InstanceID, lease.StepID)
	}
	if len(output) > 0 && !json.Valid(output) {
		return errInvalidJSON
	}
	if len(output) == 0 {
		output = json.RawMessage(`null`)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE workflow_step_attempts SET state=$4,finished_at=$5,output=$6,error_code=$7,lease_owner=NULL,lease_token=NULL,lease_expires_at=NULL
		WHERE id=$1 AND workflow_instance_id=$2 AND lease_token=$3 AND lease_epoch=$8 AND state='running' AND lease_expires_at>$5`, lease.ID.String(), lease.InstanceID.String(), lease.LeaseToken, state, at.Time(), output, emptyNull(errorCode), lease.LeaseEpoch)
	if err != nil {
		return conflict(err, "workflow="+lease.InstanceID.String()+" step="+lease.StepID)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: workflow=%s step=%s stale_lease", application.ErrConflict, lease.InstanceID, lease.StepID)
	}
	for _, m := range messages {
		if !json.Valid(m.Payload) {
			return errInvalidJSON
		}
		_, err = tx.Exec(ctx, `INSERT INTO workflow_outbox(id,workflow_instance_id,step_attempt_id,topic,payload) VALUES($1,$2,$3,$4,$5)`, m.ID, lease.InstanceID.String(), lease.ID.String(), m.Topic, m.Payload)
		if err != nil {
			return conflict(err, "workflow="+lease.InstanceID.String()+" outbox="+m.ID)
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) PutWorkflowSignal(ctx context.Context, signal application.WorkflowSignal) (bool, error) {
	if !json.Valid(signal.Payload) {
		return false, errInvalidJSON
	}
	tag, err := s.pool.Exec(ctx, `INSERT INTO workflow_signals(id,workflow_instance_id,idempotency_key,kind,payload,received_at) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(workflow_instance_id,idempotency_key) DO NOTHING`, signal.ID, signal.InstanceID.String(), signal.IdempotencyKey, signal.Kind, signal.Payload, signal.ReceivedAt.Time())
	return tag.RowsAffected() == 1, conflict(err, "workflow="+signal.InstanceID.String()+" signal="+signal.ID)
}
func (s *Store) PutWorkflowTimer(ctx context.Context, timer application.WorkflowTimer) (bool, error) {
	tag, err := s.pool.Exec(ctx, `INSERT INTO workflow_timers(id,workflow_instance_id,step_id,fire_at) VALUES($1,$2,$3,$4) ON CONFLICT(id) DO NOTHING`, timer.ID, timer.InstanceID.String(), timer.StepID, timer.FireAt.Time())
	return tag.RowsAffected() == 1, conflict(err, "workflow="+timer.InstanceID.String()+" timer="+timer.ID)
}
