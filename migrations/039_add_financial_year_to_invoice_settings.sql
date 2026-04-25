-- Migration 039: Add financial year to invoice settings for FY-based numbering

ALTER TABLE invoice_settings
ADD COLUMN IF NOT EXISTS financial_year VARCHAR(10) NOT NULL DEFAULT '';

UPDATE invoice_settings
SET financial_year = CASE
    WHEN EXTRACT(MONTH FROM created_at) >= 4
        THEN TO_CHAR(created_at, 'YY') || '-' || TO_CHAR(created_at + INTERVAL '1 year', 'YY')
    ELSE
        TO_CHAR(created_at - INTERVAL '1 year', 'YY') || '-' || TO_CHAR(created_at, 'YY')
END
WHERE TRIM(COALESCE(financial_year, '')) = '';
