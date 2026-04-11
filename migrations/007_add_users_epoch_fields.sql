-- Add epoch debug columns to users table

ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
    ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE users
SET created_at_epoch = EXTRACT(EPOCH FROM created_at)::BIGINT
WHERE created_at_epoch IS NULL;

UPDATE users
SET updated_at_epoch = EXTRACT(EPOCH FROM updated_at)::BIGINT
WHERE updated_at_epoch IS NULL;

ALTER TABLE users
    ALTER COLUMN created_at_epoch SET DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    ALTER COLUMN updated_at_epoch SET DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT;

ALTER TABLE users
    ALTER COLUMN created_at_epoch SET NOT NULL,
    ALTER COLUMN updated_at_epoch SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_created_at_epoch ON users(created_at_epoch);
CREATE INDEX IF NOT EXISTS idx_users_updated_at_epoch ON users(updated_at_epoch);
