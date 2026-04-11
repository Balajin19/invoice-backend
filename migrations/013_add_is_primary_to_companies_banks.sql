-- Add company_id FK to invoice_settings for existing databases
ALTER TABLE invoice_settings
    ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(company_id) ON DELETE CASCADE;
