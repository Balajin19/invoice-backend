-- Add HSN/SAC support for products and backfill audit emails.

ALTER TABLE IF EXISTS products
    ADD COLUMN IF NOT EXISTS hsn_sac VARCHAR(64);

-- Backfill created_by/updated_by from linked category owner where missing or placeholder.
UPDATE products p
SET
    created_by = COALESCE(NULLIF(TRIM(c.created_by), ''), NULLIF(TRIM(c.updated_by), ''), p.created_by),
    updated_by = COALESCE(NULLIF(TRIM(c.updated_by), ''), NULLIF(TRIM(c.created_by), ''), p.updated_by)
FROM categories c
WHERE c.category_id = p.category_id
  AND (
      p.created_by IS NULL OR TRIM(p.created_by) = '' OR LOWER(TRIM(p.created_by)) = 'system-migration'
      OR p.updated_by IS NULL OR TRIM(p.updated_by) = '' OR LOWER(TRIM(p.updated_by)) = 'system-migration'
  );

-- Final fallback: if still empty, use any known users.email.
UPDATE products p
SET
    created_by = COALESCE(NULLIF(TRIM(p.created_by), ''), u.email),
    updated_by = COALESCE(NULLIF(TRIM(p.updated_by), ''), u.email)
FROM LATERAL (
    SELECT email
    FROM users
    WHERE email IS NOT NULL AND TRIM(email) <> ''
    ORDER BY updated_at DESC NULLS LAST, created_at DESC NULLS LAST
    LIMIT 1
) u
WHERE p.created_by IS NULL OR TRIM(p.created_by) = '' OR p.updated_by IS NULL OR TRIM(p.updated_by) = ''
   OR LOWER(TRIM(p.created_by)) = 'system-migration' OR LOWER(TRIM(p.updated_by)) = 'system-migration';
