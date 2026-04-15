CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS units (
    unit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    created_by TEXT,
    updated_by TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_units_name_unique
    ON units (LOWER(TRIM(unit_name)));
