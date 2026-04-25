-- Migration 040: enforce one invoice_settings row per company + financial year

-- Backfill blank financial years using record creation timestamp and Apr-Mar FY logic.
UPDATE invoice_settings
SET financial_year = CASE
    WHEN EXTRACT(MONTH FROM created_at) >= 4
        THEN TO_CHAR(created_at, 'YY') || '-' || TO_CHAR(created_at + INTERVAL '1 year', 'YY')
    ELSE
        TO_CHAR(created_at - INTERVAL '1 year', 'YY') || '-' || TO_CHAR(created_at, 'YY')
END
WHERE TRIM(COALESCE(financial_year, '')) = '';

-- Keep newest row for each (company_id, financial_year) and remove older duplicates.
WITH ranked AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY company_id, LOWER(TRIM(financial_year))
            ORDER BY created_at DESC, id DESC
        ) AS rn
    FROM invoice_settings
)
DELETE FROM invoice_settings s
USING ranked r
WHERE s.id = r.id
  AND r.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoice_settings_company_fy_unique
    ON invoice_settings (company_id, LOWER(TRIM(financial_year)));
