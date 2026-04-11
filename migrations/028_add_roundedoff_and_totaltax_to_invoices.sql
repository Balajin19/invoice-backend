-- Enforce unique user emails (case-insensitive) and backfill missing audit values.

-- Normalize stored emails.
UPDATE users
SET email = lower(trim(email))
WHERE email IS NOT NULL
  AND email <> lower(trim(email));

-- Remove case-insensitive duplicates, keeping the most recently updated row.
WITH ranked AS (
    SELECT
        ctid,
        row_number() OVER (
            PARTITION BY lower(trim(email))
            ORDER BY updated_at DESC NULLS LAST, created_at DESC NULLS LAST, ctid DESC
        ) AS rn
    FROM users
    WHERE email IS NOT NULL
)
DELETE FROM users u
USING ranked r
WHERE u.ctid = r.ctid
  AND r.rn > 1;

-- Backfill audit fields when null/blank.
UPDATE users
SET
    created_by = COALESCE(NULLIF(trim(created_by), ''), lower(trim(email)), 'system-migration'),
    updated_by = COALESCE(NULLIF(trim(updated_by), ''), lower(trim(email)), COALESCE(NULLIF(trim(created_by), ''), 'system-migration')),
    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM COALESCE(created_at, NOW()))::BIGINT),
    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM COALESCE(updated_at, NOW()))::BIGINT)
WHERE created_by IS NULL
   OR trim(created_by) = ''
   OR updated_by IS NULL
   OR trim(updated_by) = ''
   OR created_at_epoch IS NULL
   OR updated_at_epoch IS NULL;

-- Enforce case-insensitive uniqueness going forward.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower_unique
ON users ((lower(email)));
