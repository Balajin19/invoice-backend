-- Replace placeholder audit values with real emails.
-- This migration is safe to run once in sequence after previous migrations.

-- 1) Invoices: prefer the linked company's audit email.
UPDATE invoices i
SET
    created_by = COALESCE(
        NULLIF(TRIM(c.created_by), ''),
        NULLIF(TRIM(c.updated_by), ''),
        i.created_by
    ),
    updated_by = COALESCE(
        NULLIF(TRIM(c.updated_by), ''),
        NULLIF(TRIM(c.created_by), ''),
        i.updated_by
    )
FROM companies c
WHERE c.company_id = i.company_id
  AND (
      LOWER(COALESCE(TRIM(i.created_by), '')) = 'system-migration'
      OR LOWER(COALESCE(TRIM(i.updated_by), '')) = 'system-migration'
      OR COALESCE(TRIM(i.created_by), '') = ''
      OR COALESCE(TRIM(i.updated_by), '') = ''
  );

-- 2) Fallback for remaining invoice rows: use latest known user email.
UPDATE invoices i
SET
    created_by = CASE
        WHEN LOWER(COALESCE(TRIM(i.created_by), '')) = 'system-migration' OR COALESCE(TRIM(i.created_by), '') = ''
            THEN u.email
        ELSE i.created_by
    END,
    updated_by = CASE
        WHEN LOWER(COALESCE(TRIM(i.updated_by), '')) = 'system-migration' OR COALESCE(TRIM(i.updated_by), '') = ''
            THEN u.email
        ELSE i.updated_by
    END
FROM LATERAL (
    SELECT email
    FROM users
    WHERE COALESCE(TRIM(email), '') <> ''
    ORDER BY updated_at DESC NULLS LAST, created_at DESC NULLS LAST
    LIMIT 1
) u
WHERE (
      LOWER(COALESCE(TRIM(i.created_by), '')) = 'system-migration'
      OR LOWER(COALESCE(TRIM(i.updated_by), '')) = 'system-migration'
      OR COALESCE(TRIM(i.created_by), '') = ''
      OR COALESCE(TRIM(i.updated_by), '') = ''
);

-- 3) Products: same fallback logic for product audit columns.
UPDATE products p
SET
    created_by = CASE
        WHEN LOWER(COALESCE(TRIM(p.created_by), '')) = 'system-migration' OR COALESCE(TRIM(p.created_by), '') = ''
            THEN u.email
        ELSE p.created_by
    END,
    updated_by = CASE
        WHEN LOWER(COALESCE(TRIM(p.updated_by), '')) = 'system-migration' OR COALESCE(TRIM(p.updated_by), '') = ''
            THEN u.email
        ELSE p.updated_by
    END
FROM LATERAL (
    SELECT email
    FROM users
    WHERE COALESCE(TRIM(email), '') <> ''
    ORDER BY updated_at DESC NULLS LAST, created_at DESC NULLS LAST
    LIMIT 1
) u
WHERE (
      LOWER(COALESCE(TRIM(p.created_by), '')) = 'system-migration'
      OR LOWER(COALESCE(TRIM(p.updated_by), '')) = 'system-migration'
      OR COALESCE(TRIM(p.created_by), '') = ''
      OR COALESCE(TRIM(p.updated_by), '') = ''
);
