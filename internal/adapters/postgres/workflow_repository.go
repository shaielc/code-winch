package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

func (s *Store) PutWorkflowDefinition(ctx context.Context, record application.WorkflowDefinitionRecord) (bool, error) {
	if !json.Valid(record.Definition) {
		return false, errInvalidJSON
	}
	tag, err := s.pool.Exec(ctx, `INSERT INTO workflow_definitions(definition_id,version,definition)
		VALUES($1,$2,$3) ON CONFLICT(definition_id,version) DO NOTHING`, record.DefinitionID, record.Version, record.Definition)
	return tag.RowsAffected() == 1, err
}

func (s *Store) CreateWorkflowInstance(ctx context.Context, instance application.WorkflowInstanceRecord, steps []application.WorkflowStepRecord, lineage *application.WorkflowLineageRecord) error {
	if !json.Valid(instance.Inputs) {
		return errInvalidJSON
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, `INSERT INTO workflow_instances(id,definition_id,definition_version,status,inputs,created_at)
		VALUES($1,$2,$3,$4,$5,$6)`, instance.ID.String(), instance.DefinitionID, instance.DefinitionVersion, instance.Status, instance.Inputs, instance.CreatedAt.Time())
	if err != nil {
		return conflict(err, "workflow="+instance.ID.String())
	}
	for _, step := range steps {
		if step.InstanceID != instance.ID || step.Attempt != 1 || step.State != "ready" {
			return fmt.Errorf("%w: workflow=%s invalid_initial_attempt", application.ErrConflict, instance.ID)
		}
		if err = insertWorkflowAttempt(ctx, tx, step); err != nil {
			return conflict(err, "workflow="+instance.ID.String()+" step="+step.StepID)
		}
	}
	if lineage != nil {
		if lineage.InstanceID != instance.ID {
			return fmt.Errorf("%w: workflow=%s invalid_lineage", application.ErrConflict, instance.ID)
		}
		_, err = tx.Exec(ctx, `INSERT INTO workflow_lineage(instance_id,parent_instance_id,parent_step_id) VALUES($1,$2,$3)`, instance.ID.String(), lineage.ParentInstanceID.String(), lineage.ParentStepID)
		if err != nil {
			return conflict(err, "workflow="+instance.ID.String())
		}
	}
	return tx.Commit(ctx)
}

func insertWorkflowAttempt(ctx context.Context, tx pgx.Tx, step application.WorkflowStepRecord) error {
	_, err := tx.Exec(ctx, `INSERT INTO workflow_step_attempts(instance_id,step_id,attempt,attempt_id,state,available_at)
		VALUES($1,$2,$3,$4,$5,$6)`, step.InstanceID.String(), step.StepID, step.Attempt, step.AttemptID.String(), step.State, step.AvailableAt.Time())
	return err
}

type timestampScanner struct{ target *domain.Timestamp }

func (s timestampScanner) Scan(value any) error {
	v, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("postgres repository: invalid timestamp type")
	}
	ts, err := domain.NewTimestamp(v)
	if err == nil {
		*s.target = ts
	}
	return err
}

func (s *Store) AppendWorkflowAttempt(ctx context.Context, step application.WorkflowStepRecord) error {
	if step.Attempt < 2 || step.State != "ready" {
		return fmt.Errorf("%w: workflow=%s step=%s invalid_attempt_sequence", application.ErrConflict, step.InstanceID, step.StepID)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var previousState string
	err = tx.QueryRow(ctx, `SELECT state FROM workflow_step_attempts WHERE instance_id=$1 AND step_id=$2 AND attempt=$3 FOR UPDATE`, step.InstanceID.String(), step.StepID, step.Attempt-1).Scan(&previousState)
	if errors.Is(err, pgx.ErrNoRows) || (previousState != "failed" && previousState != "cancelled") {
		return fmt.Errorf("%w: workflow=%s step=%s invalid_attempt_sequence", application.ErrConflict, step.InstanceID, step.StepID)
	}
	if err != nil {
		return err
	}
	if err = insertWorkflowAttempt(ctx, tx, step); err != nil {
		return conflict(err, "workflow="+step.InstanceID.String()+" step="+step.StepID)
	}
	return tx.Commit(ctx)
}

func (s *Store) GetWorkflowInstance(ctx context.Context, id domain.WorkflowID) (application.WorkflowInstanceRecord, []application.WorkflowStepRecord, error) {
	record := application.WorkflowInstanceRecord{ID: id}
	err := s.pool.QueryRow(ctx, `SELECT definition_id,definition_version,status,inputs,created_at FROM workflow_instances WHERE id=$1`, id.String()).Scan(&record.DefinitionID, &record.DefinitionVersion, &record.Status, &record.Inputs, timestampScanner{target: &record.CreatedAt})
	if errors.Is(err, pgx.ErrNoRows) {
		return record, nil, application.ErrNotFound
	}
	if err != nil {
		return record, nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT step_id,attempt,attempt_id,state,available_at FROM workflow_step_attempts WHERE instance_id=$1 ORDER BY step_id,attempt`, id.String())
	if err != nil {
		return record, nil, err
	}
	defer rows.Close()
	steps := []application.WorkflowStepRecord{}
	for rows.Next() {
		step := application.WorkflowStepRecord{InstanceID: id}
		var attemptID string
		var availableAt timestampScanner
		availableAt.target = &step.AvailableAt
		if err = rows.Scan(&step.StepID, &step.Attempt, &attemptID, &step.State, &availableAt); err != nil {
			return record, nil, err
		}
		step.AttemptID, _ = domain.ParseAttemptID(attemptID)
		steps = append(steps, step)
	}
	return record, steps, rows.Err()
}

func (s *Store) ClaimReadySteps(ctx context.Context, owner, token string, now, until domain.Timestamp, limit int) ([]application.StepLease, error) {
	if limit <= 0 {
		return []application.StepLease{}, nil
	}
	rows, err := s.pool.Query(ctx, `WITH candidates AS (
		SELECT instance_id,step_id,attempt FROM workflow_step_attempts
		WHERE available_at <= $3 AND (state='ready' OR (state='leased' AND lease_expires_at <= $3))
		ORDER BY available_at,instance_id,step_id,attempt FOR UPDATE SKIP LOCKED LIMIT $4
	) UPDATE workflow_step_attempts a SET state='leased',lease_owner=$1,lease_token=$2,lease_expires_at=$5,
		started_at=COALESCE(started_at,$3) FROM candidates c
	WHERE a.instance_id=c.instance_id AND a.step_id=c.step_id AND a.attempt=c.attempt
	RETURNING a.instance_id::text,a.step_id,a.attempt,a.attempt_id::text,a.available_at,a.lease_expires_at`, owner, token, now.Time(), limit, until.Time())
	if err != nil {
		return nil, conflict(err, "workflow_step_claim")
	}
	defer rows.Close()
	leases := []application.StepLease{}
	for rows.Next() {
		lease := application.StepLease{LeaseOwner: owner, LeaseToken: token}
		var instanceID, attemptID string
		var availableAt, expiresAt timestampScanner
		availableAt.target, expiresAt.target = &lease.AvailableAt, &lease.LeaseExpiresAt
		if err = rows.Scan(&instanceID, &lease.StepID, &lease.Attempt, &attemptID, &availableAt, &expiresAt); err != nil {
			return nil, err
		}
		lease.InstanceID, _ = domain.ParseWorkflowID(instanceID)
		lease.AttemptID, _ = domain.ParseAttemptID(attemptID)
		lease.State = "leased"
		leases = append(leases, lease)
	}
	return leases, rows.Err()
}

func (s *Store) CompleteStep(ctx context.Context, instanceID domain.WorkflowID, stepID string, attempt uint, token, state string, at domain.Timestamp, output []byte) error {
	if (state != "completed" && state != "failed" && state != "cancelled") || (len(output) > 0 && !json.Valid(output)) {
		return fmt.Errorf("%w: workflow=%s step=%s invalid_completion", application.ErrConflict, instanceID, stepID)
	}
	if len(output) == 0 {
		output = []byte("null")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE workflow_step_attempts SET state=$5,finished_at=$6,output=$7,
		lease_owner=NULL,lease_token=NULL,lease_expires_at=NULL
		WHERE instance_id=$1 AND step_id=$2 AND attempt=$3 AND state='leased' AND lease_token=$4`, instanceID.String(), stepID, attempt, token, state, at.Time(), output)
	if err != nil {
		return conflict(err, "workflow="+instanceID.String()+" step="+stepID)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: workflow=%s step=%s lease_token_stale", application.ErrConflict, instanceID, stepID)
	}
	payload, _ := json.Marshal(map[string]any{"workflowId": instanceID.String(), "stepId": stepID, "attempt": attempt, "state": state})
	_, err = tx.Exec(ctx, `INSERT INTO workflow_outbox(id,instance_id,topic,payload)
		SELECT attempt_id,instance_id,'workflow.step.finished',$2 FROM workflow_step_attempts
		WHERE instance_id=$1 AND step_id=$3 AND attempt=$4`, instanceID.String(), payload, stepID, attempt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) PutWorkflowSignal(ctx context.Context, signal application.WorkflowSignalRecord) (bool, error) {
	if !json.Valid(signal.Payload) {
		return false, errInvalidJSON
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `INSERT INTO workflow_signals(id,instance_id,idempotency_key,kind,payload,received_at)
		VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(instance_id,idempotency_key) DO NOTHING`, signal.ID.String(), signal.InstanceID.String(), signal.IdempotencyKey, signal.Kind, signal.Payload, signal.ReceivedAt.Time())
	if err != nil || tag.RowsAffected() == 0 {
		return false, conflict(err, "workflow="+signal.InstanceID.String())
	}
	payload, _ := json.Marshal(map[string]string{"workflowId": signal.InstanceID.String(), "signalId": signal.ID.String(), "kind": signal.Kind})
	if _, err = tx.Exec(ctx, `INSERT INTO workflow_outbox(id,instance_id,topic,payload) VALUES($1,$2,'workflow.signal.received',$3)`, signal.ID.String(), signal.InstanceID.String(), payload); err != nil {
		return false, conflict(err, "workflow="+signal.InstanceID.String())
	}
	return true, tx.Commit(ctx)
}

func (s *Store) PutWorkflowTimer(ctx context.Context, timer application.WorkflowTimerRecord) (bool, error) {
	tag, err := s.pool.Exec(ctx, `INSERT INTO workflow_timers(id,instance_id,step_id,fire_at) VALUES($1,$2,$3,$4) ON CONFLICT(id) DO NOTHING`, timer.ID.String(), timer.InstanceID.String(), timer.StepID, timer.FireAt.Time())
	return tag.RowsAffected() == 1, conflict(err, "workflow="+timer.InstanceID.String())
}

func (s *Store) FireWorkflowTimers(ctx context.Context, now domain.Timestamp, limit int) ([]application.WorkflowTimerRecord, error) {
	if limit <= 0 {
		return []application.WorkflowTimerRecord{}, nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, `WITH ready AS (SELECT id FROM workflow_timers WHERE fired_at IS NULL AND fire_at <= $1 ORDER BY fire_at FOR UPDATE SKIP LOCKED LIMIT $2)
		UPDATE workflow_timers t SET fired_at=$1 FROM ready r WHERE t.id=r.id RETURNING t.id::text,t.instance_id::text,t.step_id,t.fire_at`, now.Time(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	timers := []application.WorkflowTimerRecord{}
	for rows.Next() {
		var timer application.WorkflowTimerRecord
		var id, instanceID string
		var fireAt timestampScanner
		fireAt.target = &timer.FireAt
		if err = rows.Scan(&id, &instanceID, &timer.StepID, &fireAt); err != nil {
			return nil, err
		}
		timer.ID, _ = domain.ParseCommandID(id)
		timer.InstanceID, _ = domain.ParseWorkflowID(instanceID)
		timers = append(timers, timer)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	for _, timer := range timers {
		payload, _ := json.Marshal(map[string]string{"workflowId": timer.InstanceID.String(), "timerId": timer.ID.String(), "stepId": timer.StepID})
		if _, err = tx.Exec(ctx, `INSERT INTO workflow_outbox(id,instance_id,topic,payload) VALUES($1,$2,'workflow.timer.fired',$3)`, timer.ID.String(), timer.InstanceID.String(), payload); err != nil {
			return nil, conflict(err, "workflow="+timer.InstanceID.String())
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return timers, nil
}
