CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Ensure dependency table exists for safe FK creation in fresh environments.
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

-- Ensure customer_products table exists in case this migration is run independently.
CREATE TABLE IF NOT EXISTS customer_products (
    id BIGSERIAL PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(customer_id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(product_id) ON DELETE CASCADE,
    unit_id UUID,
    price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    created_by TEXT,
    updated_by TEXT,
    UNIQUE (customer_id, product_id)
);

ALTER TABLE customer_products
    ADD COLUMN IF NOT EXISTS unit_id UUID;

UPDATE customer_products cp
SET unit_id = p.unit_id
FROM products p
WHERE cp.product_id = p.product_id
  AND cp.unit_id IS NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_customer_products_unit_id'
    ) THEN
        ALTER TABLE customer_products
            ADD CONSTRAINT fk_customer_products_unit_id
            FOREIGN KEY (unit_id) REFERENCES units(unit_id) ON DELETE RESTRICT;
    END IF;
END
$$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM customer_products WHERE unit_id IS NULL) THEN
        RAISE NOTICE 'customer_products.unit_id contains NULL values; NOT NULL constraint not applied';
    ELSE
        ALTER TABLE customer_products
            ALTER COLUMN unit_id SET NOT NULL;
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_customer_products_unit_id ON customer_products(unit_id);