-- Add owner_name column to companies for existing databases
ALTER TABLE companies
    ADD COLUMN IF NOT EXISTS owner_name VARCHAR(255);