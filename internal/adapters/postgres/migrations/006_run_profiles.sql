-- migrate:up
ALTER TABLE runs
  ADD COLUMN workspace_path text NOT NULL DEFAULT '',
  ADD COLUMN harness_profile text NOT NULL DEFAULT '',
  ADD COLUMN sandbox_profile text NOT NULL DEFAULT '',
  ADD COLUMN created_at timestamptz NOT NULL DEFAULT transaction_timestamp(),
  ADD COLUMN updated_at timestamptz NOT NULL DEFAULT transaction_timestamp();

-- migrate:down
ALTER TABLE runs DROP COLUMN updated_at, DROP COLUMN created_at, DROP COLUMN sandbox_profile, DROP COLUMN harness_profile, DROP COLUMN workspace_path;
