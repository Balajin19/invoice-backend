-- Add IGST support to invoices for inter-state tax calculation.

ALTER TABLE IF EXISTS invoices
    ADD COLUMN IF NOT EXISTS igst NUMERIC(12, 2) DEFAULT 0;

UPDATE invoices
SET igst = 0
WHERE igst IS NULL;
