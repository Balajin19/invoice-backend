-- Add payment_terms column to invoices for existing databases
ALTER TABLE invoices
    ADD COLUMN IF NOT EXISTS payment_terms TEXT;
