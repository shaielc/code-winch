-- migrate:up
CREATE TABLE workflow_definitions (
  definition_id text NOT NULL,
  version text NOT NULL,
  document jsonb NOT NULL,
  created_at timestamptz NOT NULL,
  PRIMARY KEY (definition_id, version)
);

CREATE TABLE workflow_instances (
  id uuid PRIMARY KEY,
  definition_id text NOT NULL,
  definition_version text NOT NULL,
  status text NOT NULL CHECK (status IN ('pending','running','completed','failed','cancelled')),
  inputs jsonb NOT NULL DEFAULT '{}'::jsonb,
  harness_profile text NOT NULL,
  sandbox_profile text NOT NULL,
  parent_instance_id uuid REFERENCES workflow_instances(id),
  parent_step_id text,
  created_at timestamptz NOT NULL,
  FOREIGN KEY (definition_id, definition_version) REFERENCES workflow_definitions(definition_id, version),
  CHECK ((parent_instance_id IS NULL) = (parent_step_id IS NULL))
);

CREATE TABLE workflow_step_attempts (
  id uuid PRIMARY KEY,
  workflow_instance_id uuid NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  step_id text NOT NULL,
  attempt integer NOT NULL CHECK (attempt > 0),
  state text NOT NULL CHECK (state IN ('ready','running','completed','failed','cancelled')),
  available_at timestamptz NOT NULL,
  started_at timestamptz,
  finished_at timestamptz,
  output jsonb,
  error_code text,
  lease_owner text,
  lease_token uuid,
  lease_epoch bigint NOT NULL DEFAULT 0 CHECK (lease_epoch >= 0),
  lease_expires_at timestamptz,
  UNIQUE (workflow_instance_id, step_id, attempt)
);
CREATE INDEX workflow_step_claim_idx ON workflow_step_attempts (available_at, workflow_instance_id, step_id)
  WHERE state IN ('ready','running');

CREATE FUNCTION protect_terminal_workflow_attempt() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF OLD.state IN ('completed','failed','cancelled') AND NEW IS DISTINCT FROM OLD THEN
    RAISE EXCEPTION 'terminal workflow attempt history is immutable' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END $$;
CREATE TRIGGER workflow_attempt_terminal_immutable BEFORE UPDATE ON workflow_step_attempts
  FOR EACH ROW EXECUTE FUNCTION protect_terminal_workflow_attempt();

CREATE TABLE workflow_timers (
  id uuid PRIMARY KEY,
  workflow_instance_id uuid NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  step_id text NOT NULL,
  fire_at timestamptz NOT NULL,
  consumed_at timestamptz
);
CREATE INDEX workflow_timer_ready_idx ON workflow_timers (fire_at) WHERE consumed_at IS NULL;

CREATE TABLE workflow_signals (
  id uuid PRIMARY KEY,
  workflow_instance_id uuid NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  idempotency_key text NOT NULL,
  kind text NOT NULL,
  payload jsonb NOT NULL,
  received_at timestamptz NOT NULL,
  consumed_at timestamptz,
  UNIQUE (workflow_instance_id, idempotency_key)
);

CREATE TABLE workflow_outbox (
  id uuid PRIMARY KEY,
  workflow_instance_id uuid NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  step_attempt_id uuid NOT NULL REFERENCES workflow_step_attempts(id),
  topic text NOT NULL,
  payload jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT transaction_timestamp(),
  delivered_at timestamptz
);

-- migrate:down
DROP TABLE workflow_outbox;
DROP TABLE workflow_signals;
DROP TABLE workflow_timers;
DROP TRIGGER workflow_attempt_terminal_immutable ON workflow_step_attempts;
DROP FUNCTION protect_terminal_workflow_attempt();
DROP TABLE workflow_step_attempts;
DROP TABLE workflow_instances;
DROP TABLE workflow_definitions;
