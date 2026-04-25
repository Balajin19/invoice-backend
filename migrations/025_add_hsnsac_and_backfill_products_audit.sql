-- Store rounded off and total tax values on invoices.

ALTER TABLE IF EXISTS invoices
    ADD COLUMN IF NOT EXISTS rounded_off NUMERIC(12, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_tax NUMERIC(12, 2) DEFAULT 0;

UPDATE invoices
SET
    rounded_off = COALESCE(rounded_off, 0),
    total_tax = COALESCE(total_tax, COALESCE(cgst, 0) + COALESCE(sgst, 0) + COALESCE(igst, 0), 0)
WHERE rounded_off IS NULL
   OR total_tax IS NULL;