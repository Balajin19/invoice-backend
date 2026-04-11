-- Link invoices to companies so invoice data can be scoped by selected company
ALTER TABLE invoices
    ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(company_id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_invoices_company_id ON invoices(company_id);
