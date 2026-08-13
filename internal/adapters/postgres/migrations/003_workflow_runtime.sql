-- migrate:up
CREATE TABLE workflow_definitions (
  definition_id text NOT NULL,
  version text NOT NULL,
  definition jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT transaction_timestamp(),
  PRIMARY KEY (definition_id, version)
);
CREATE TABLE workflow_instances (
  id uuid PRIMARY KEY,
  definition_id text NOT NULL,
  definition_version text NOT NULL,
  status text NOT NULL CHECK (status IN ('pending','running','completed','failed','cancelled')),
  inputs jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL,
  completed_at timestamptz,
  FOREIGN KEY (definition_id, definition_version)
    REFERENCES workflow_definitions(definition_id, version)
);
CREATE TABLE workflow_lineage (
  instance_id uuid PRIMARY KEY REFERENCES workflow_instances(id) ON DELETE CASCADE,
  parent_instance_id uuid NOT NULL REFERENCES workflow_instances(id),
  parent_step_id text NOT NULL
);
CREATE TABLE workflow_step_attempts (
  instance_id uuid NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  step_id text NOT NULL,
  attempt integer NOT NULL CHECK (attempt > 0),
  attempt_id uuid NOT NULL UNIQUE,
  state text NOT NULL CHECK (state IN ('ready','leased','completed','failed','cancelled')),
  available_at timestamptz NOT NULL,
  lease_owner text,
  lease_token uuid,
  lease_expires_at timestamptz,
  started_at timestamptz,
  finished_at timestamptz,
  output jsonb,
  PRIMARY KEY (instance_id, step_id, attempt),
  CHECK ((state = 'leased') = (lease_owner IS NOT NULL AND lease_token IS NOT NULL AND lease_expires_at IS NOT NULL))
);
CREATE INDEX workflow_step_claim_idx ON workflow_step_attempts (available_at, instance_id, step_id, attempt)
  WHERE state IN ('ready','leased');
CREATE FUNCTION protect_workflow_attempt_history() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF OLD.state IN ('completed','failed','cancelled') AND NEW IS DISTINCT FROM OLD THEN
    RAISE EXCEPTION 'terminal workflow attempt history is immutable' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END $$;
CREATE TRIGGER workflow_attempt_terminal_immutable BEFORE UPDATE ON workflow_step_attempts
  FOR EACH ROW EXECUTE FUNCTION protect_workflow_attempt_history();
CREATE TABLE workflow_signals (
  id uuid PRIMARY KEY,
  instance_id uuid NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  idempotency_key text NOT NULL,
  kind text NOT NULL,
  payload jsonb NOT NULL,
  received_at timestamptz NOT NULL,
  UNIQUE (instance_id, idempotency_key)
);
CREATE TABLE workflow_timers (
  id uuid PRIMARY KEY,
  instance_id uuid NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  step_id text NOT NULL,
  fire_at timestamptz NOT NULL,
  fired_at timestamptz
);
CREATE INDEX workflow_timer_ready_idx ON workflow_timers (fire_at) WHERE fired_at IS NULL;
CREATE TABLE workflow_outbox (
  id uuid PRIMARY KEY,
  instance_id uuid NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  topic text NOT NULL,
  payload jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT transaction_timestamp()
);

-- migrate:down
DROP TABLE workflow_outbox;
DROP TABLE workflow_timers;
DROP TABLE workflow_signals;
DROP TRIGGER workflow_attempt_terminal_immutable ON workflow_step_attempts;
DROP FUNCTION protect_workflow_attempt_history();
DROP TABLE workflow_step_attempts;
DROP TABLE workflow_lineage;
DROP TABLE workflow_instances;
DROP TABLE workflow_definitions;
