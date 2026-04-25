-- Add soft-delete flag to invoices for existing databases
ALTER TABLE invoices
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;

UPDATE invoices
SET is_active = TRUE
WHERE is_active IS NULL;