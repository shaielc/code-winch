-- migrate:up
CREATE TABLE runs (
  id uuid PRIMARY KEY,
  version bigint NOT NULL CHECK (version > 0),
  resolved_configuration jsonb NOT NULL DEFAULT '{}'::jsonb,
  last_sequence bigint NOT NULL DEFAULT 0 CHECK (last_sequence >= 0)
);
CREATE TABLE run_attempts (
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  ordinal integer NOT NULL CHECK (ordinal > 0),
  id uuid NOT NULL UNIQUE,
  previous_attempt_id uuid,
  state text NOT NULL CHECK (state IN ('created','queued','preparing','running','stopping','completed','failed','cancelled')),
  PRIMARY KEY (run_id, ordinal)
);
CREATE FUNCTION protect_terminal_attempt() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF OLD.state IN ('completed','failed','cancelled') AND NEW IS DISTINCT FROM OLD THEN
    RAISE EXCEPTION 'terminal attempt history is immutable' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END $$;
CREATE TRIGGER run_attempt_terminal_immutable BEFORE UPDATE ON run_attempts
  FOR EACH ROW EXECUTE FUNCTION protect_terminal_attempt();
CREATE TABLE run_events (
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  sequence bigint NOT NULL CHECK (sequence > 0),
  event_id uuid NOT NULL UNIQUE,
  occurred_at timestamptz NOT NULL,
  kind text NOT NULL,
  schema_version integer NOT NULL CHECK (schema_version > 0),
  source jsonb NOT NULL,
  sensitivity text NOT NULL CHECK (sensitivity <> 'secret'),
  payload jsonb NOT NULL,
  extensions jsonb NOT NULL DEFAULT '{}'::jsonb,
  PRIMARY KEY (run_id, sequence)
);
CREATE TABLE input_commands (
  id uuid PRIMARY KEY,
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  idempotency_key text NOT NULL,
  kind text NOT NULL,
  payload jsonb NOT NULL,
  UNIQUE (run_id, idempotency_key)
);

-- migrate:down
DROP TABLE input_commands;
DROP TABLE run_events;
DROP TRIGGER run_attempt_terminal_immutable ON run_attempts;
DROP FUNCTION protect_terminal_attempt();
DROP TABLE run_attempts;
DROP TABLE runs;
