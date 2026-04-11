ALTER TABLE IF EXISTS banks
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE banks
SET
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);
