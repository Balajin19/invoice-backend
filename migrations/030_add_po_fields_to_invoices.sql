-- Rename primary key column from id to company_id for existing databases
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'companies'
          AND column_name = 'id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'companies'
          AND column_name = 'company_id'
    ) THEN
        ALTER TABLE companies RENAME COLUMN id TO company_id;
    END IF;
END $$;
