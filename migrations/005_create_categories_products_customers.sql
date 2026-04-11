CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- Missing core entity tables used by repositories.
CREATE TABLE IF NOT EXISTS categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    created_by TEXT,
    updated_by TEXT
);

CREATE TABLE IF NOT EXISTS products (
    product_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_name VARCHAR(255) NOT NULL,
    hsn_sac VARCHAR(50),
    unit VARCHAR(100) NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(category_id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    created_by TEXT,
    updated_by TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_name_unit_category_unique
    ON products (LOWER(TRIM(product_name)), LOWER(TRIM(unit)), category_id);

CREATE TABLE IF NOT EXISTS customers (
    customer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_name VARCHAR(255) NOT NULL,
    building_number VARCHAR(100) NOT NULL,
    street VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    district VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    pincode VARCHAR(20) NOT NULL,
    gstin VARCHAR(20),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    created_by TEXT,
    updated_by TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_identity_unique
    ON customers (
        LOWER(TRIM(customer_name)),
        LOWER(TRIM(building_number)),
        LOWER(TRIM(COALESCE(street, ''))),
        LOWER(TRIM(city)),
        LOWER(TRIM(district)),
        LOWER(TRIM(state)),
        LOWER(TRIM(pincode))
    );

CREATE TABLE IF NOT EXISTS customer_products (
    id BIGSERIAL PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(customer_id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(product_id) ON DELETE CASCADE,
    price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at_epoch BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    created_by TEXT,
    updated_by TEXT,
    UNIQUE (customer_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_customer_products_customer_id ON customer_products(customer_id);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
