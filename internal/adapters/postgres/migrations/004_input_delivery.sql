-- migrate:up
ALTER TABLE runs ADD COLUMN input_command_sequence bigint NOT NULL DEFAULT 0 CHECK (input_command_sequence >= 0);

ALTER TABLE input_commands
  ADD COLUMN actor_id text NOT NULL DEFAULT '',
  ADD COLUMN expected_state text,
  ADD COLUMN expected_sequence bigint,
  ADD COLUMN accepted_at timestamptz NOT NULL DEFAULT transaction_timestamp();

ALTER TABLE outbox DROP CONSTRAINT outbox_run_id_sequence_key;
ALTER TABLE outbox ADD CONSTRAINT outbox_run_topic_sequence_key UNIQUE (run_id, topic, sequence);

-- migrate:down
ALTER TABLE outbox DROP CONSTRAINT outbox_run_topic_sequence_key;
ALTER TABLE outbox ADD CONSTRAINT outbox_run_id_sequence_key UNIQUE (run_id, sequence);
ALTER TABLE input_commands DROP COLUMN accepted_at, DROP COLUMN expected_sequence, DROP COLUMN expected_state, DROP COLUMN actor_id;
ALTER TABLE runs DROP COLUMN input_command_sequence;
