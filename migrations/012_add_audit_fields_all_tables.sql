-- Standardize audit columns across all tables used by this service.
-- Adds missing created_at, updated_at, created_by, updated_by, created_at_epoch, updated_at_epoch.
-- Backfills values for existing rows.

-- invoices
ALTER TABLE IF EXISTS invoices
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE invoices
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- users
ALTER TABLE IF EXISTS users
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE users
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- login_sessions
ALTER TABLE IF EXISTS login_sessions
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE login_sessions
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- password_reset_tokens
ALTER TABLE IF EXISTS password_reset_tokens
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE password_reset_tokens
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- companies
ALTER TABLE IF EXISTS companies
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE companies
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- banks
ALTER TABLE IF EXISTS banks
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE banks
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- invoice_settings
ALTER TABLE IF EXISTS invoice_settings
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE invoice_settings
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- customers
ALTER TABLE IF EXISTS customers
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE customers
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- products
ALTER TABLE IF EXISTS products
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE products
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- categories
ALTER TABLE IF EXISTS categories
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE categories
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- invoice_products
ALTER TABLE IF EXISTS invoice_products
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE invoice_products
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);

-- customer_products
ALTER TABLE IF EXISTS customer_products
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_by TEXT,
ADD COLUMN IF NOT EXISTS updated_by TEXT,
ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT,
ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT;

UPDATE customer_products
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW()),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT);
