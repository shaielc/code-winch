-- migrate:up
CREATE TABLE outbox (
  id uuid PRIMARY KEY,
  topic text NOT NULL,
  payload jsonb NOT NULL,
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  sequence bigint NOT NULL,
  available_at timestamptz NOT NULL DEFAULT transaction_timestamp(),
  attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  lease_owner text,
  lease_token uuid,
  lease_expires_at timestamptz,
  completed_at timestamptz,
  poisoned_at timestamptz,
  last_error_code text,
  UNIQUE (run_id, sequence)
);
CREATE INDEX outbox_claim_idx ON outbox (available_at, run_id, sequence)
  WHERE completed_at IS NULL AND poisoned_at IS NULL;

-- migrate:down
DROP TABLE outbox;
