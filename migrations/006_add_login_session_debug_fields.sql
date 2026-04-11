-- Add timezone and epoch debug columns to login_sessions

ALTER TABLE IF EXISTS login_sessions
    ADD COLUMN IF NOT EXISTS login_timezone TEXT,
    ADD COLUMN IF NOT EXISTS logged_in_epoch BIGINT,
    ADD COLUMN IF NOT EXISTS expires_in_epoch BIGINT;

CREATE INDEX IF NOT EXISTS idx_login_sessions_logged_in_epoch ON login_sessions(logged_in_epoch);
CREATE INDEX IF NOT EXISTS idx_login_sessions_expires_in_epoch ON login_sessions(expires_in_epoch);
