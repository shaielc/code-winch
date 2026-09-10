-- migrate:up
-- A run's own fields are columns. resolved_configuration carries the fully
-- resolved non-secret configuration (docs/code-structure.md section 6) and its
-- writer replaces it as one document, so nothing else may live in it.
ALTER TABLE runs
  ADD COLUMN workspace_path text NOT NULL DEFAULT '',
  ADD COLUMN harness_profile text NOT NULL DEFAULT '',
  ADD COLUMN sandbox_profile text NOT NULL DEFAULT '',
  ADD COLUMN created_at timestamptz NOT NULL DEFAULT transaction_timestamp(),
  ADD COLUMN updated_at timestamptz NOT NULL DEFAULT transaction_timestamp();

-- Idempotent creation follows input_commands and workflow_signals: the key is a
-- column on the row the request creates, and the database enforces uniqueness.
-- A run no client request created leaves both NULL, which a unique index treats
-- as distinct, so internal creation is not serialized behind one key.
ALTER TABLE runs
  ADD COLUMN create_actor text,
  ADD COLUMN create_idempotency_key text;
ALTER TABLE runs ADD CONSTRAINT runs_create_request_key
  UNIQUE (create_actor, create_idempotency_key);
ALTER TABLE runs ADD CONSTRAINT runs_create_request_complete
  CHECK ((create_actor IS NULL) = (create_idempotency_key IS NULL));

-- migrate:down
ALTER TABLE runs DROP CONSTRAINT runs_create_request_complete;
ALTER TABLE runs DROP CONSTRAINT runs_create_request_key;
ALTER TABLE runs
  DROP COLUMN create_idempotency_key,
  DROP COLUMN create_actor,
  DROP COLUMN updated_at,
  DROP COLUMN created_at,
  DROP COLUMN sandbox_profile,
  DROP COLUMN harness_profile,
  DROP COLUMN workspace_path;
