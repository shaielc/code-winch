-- migrate:up
ALTER TABLE runs
  ADD COLUMN desired_state text,
  ADD COLUMN harness_driver text NOT NULL DEFAULT '',
  ADD COLUMN sandbox_driver text NOT NULL DEFAULT '',
  ADD COLUMN execution_id text NOT NULL DEFAULT '',
  ADD COLUMN supervisor_lease_owner text,
  ADD COLUMN supervisor_lease_token uuid,
  ADD COLUMN supervisor_lease_epoch bigint NOT NULL DEFAULT 0 CHECK (supervisor_lease_epoch >= 0),
  ADD COLUMN supervisor_lease_expires_at timestamptz,
  ADD COLUMN last_runner_ordinal bigint NOT NULL DEFAULT 0 CHECK (last_runner_ordinal >= 0);

-- migrate:down
ALTER TABLE runs
  DROP COLUMN last_runner_ordinal,
  DROP COLUMN supervisor_lease_expires_at,
  DROP COLUMN supervisor_lease_epoch,
  DROP COLUMN supervisor_lease_token,
  DROP COLUMN supervisor_lease_owner,
  DROP COLUMN execution_id,
  DROP COLUMN sandbox_driver,
  DROP COLUMN harness_driver,
  DROP COLUMN desired_state;
