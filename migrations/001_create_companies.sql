-- Companies settings table

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS companies (
    company_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    company_name VARCHAR(255) NOT NULL,
    owner_name VARCHAR(255),
    building_number VARCHAR(50),
    street VARCHAR(255),
    city VARCHAR(100),
    district VARCHAR(100),
    state VARCHAR(100),
    pincode VARCHAR(10),

    gstin VARCHAR(20),
    cgst_rate NUMERIC(10, 2) DEFAULT 0,
    sgst_rate NUMERIC(10, 2) DEFAULT 0,
    igst_rate NUMERIC(10, 2) DEFAULT 0,
    email TEXT,
    mobile_number VARCHAR(20),
    is_primary BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    created_by TEXT,
    updated_by TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_created_by_unique ON companies(lower(created_by));
CREATE INDEX IF NOT EXISTS idx_companies_updated_at ON companies(updated_at);
