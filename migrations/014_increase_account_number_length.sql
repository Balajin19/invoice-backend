-- Add or align company-level tax-rate fields for existing databases
ALTER TABLE companies
    ADD COLUMN IF NOT EXISTS cgst_rate NUMERIC(10, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sgst_rate NUMERIC(10, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS igst_rate NUMERIC(10, 2) DEFAULT 0;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'companies' AND column_name = 'cgst'
    ) THEN
        UPDATE companies
        SET cgst_rate = COALESCE(cgst_rate, cgst, 0);
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'companies' AND column_name = 'sgst'
    ) THEN
        UPDATE companies
        SET sgst_rate = COALESCE(sgst_rate, sgst, 0);
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'companies' AND column_name = 'igst'
    ) THEN
        UPDATE companies
        SET igst_rate = COALESCE(igst_rate, igst, 0);
    END IF;
END $$;