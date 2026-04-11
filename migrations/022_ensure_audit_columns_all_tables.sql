-- Ensure standard audit columns exist on all application tables
-- Columns: created_at, updated_at, created_by, updated_by, created_at_epoch, updated_at_epoch
-- Safe to run multiple times.

DO $$
DECLARE
    t RECORD;
    has_email_column BOOLEAN;
BEGIN
    FOR t IN
        SELECT tablename
        FROM pg_tables
        WHERE schemaname = 'public'
          AND tablename <> 'schema_migrations'
    LOOP
        EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW()', t.tablename);
        EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW()', t.tablename);
        EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS created_by TEXT', t.tablename);
        EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS updated_by TEXT', t.tablename);
        EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS created_at_epoch BIGINT', t.tablename);
        EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS updated_at_epoch BIGINT', t.tablename);

        SELECT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = t.tablename
              AND column_name = 'email'
        ) INTO has_email_column;

        IF has_email_column THEN
            EXECUTE format($sql$
                UPDATE %I
                SET
                    created_at = COALESCE(created_at, NOW()),
                    updated_at = COALESCE(updated_at, NOW()),
                    created_by = COALESCE(created_by, NULLIF(TRIM(email::text), ''), 'system-migration'),
                    updated_by = COALESCE(updated_by, NULLIF(TRIM(email::text), ''), 'system-migration'),
                    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
                    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT)
                WHERE
                    created_at IS NULL
                    OR updated_at IS NULL
                    OR created_by IS NULL
                    OR updated_by IS NULL
                    OR created_at_epoch IS NULL
                    OR updated_at_epoch IS NULL
            $sql$, t.tablename);
        ELSE
            EXECUTE format($sql$
                UPDATE %I
                SET
                    created_at = COALESCE(created_at, NOW()),
                    updated_at = COALESCE(updated_at, NOW()),
                    created_by = COALESCE(created_by, 'system-migration'),
                    updated_by = COALESCE(updated_by, 'system-migration'),
                    created_at_epoch = COALESCE(created_at_epoch, EXTRACT(EPOCH FROM created_at)::BIGINT),
                    updated_at_epoch = COALESCE(updated_at_epoch, EXTRACT(EPOCH FROM updated_at)::BIGINT)
                WHERE
                    created_at IS NULL
                    OR updated_at IS NULL
                    OR created_by IS NULL
                    OR updated_by IS NULL
                    OR created_at_epoch IS NULL
                    OR updated_at_epoch IS NULL
            $sql$, t.tablename);
        END IF;
    END LOOP;
END $$;
