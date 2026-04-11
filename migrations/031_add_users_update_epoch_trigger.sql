-- Keep updated_at and updated_at_epoch in sync on row updates for core tables.

CREATE OR REPLACE FUNCTION set_updated_epoch_fields()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    NEW.updated_at_epoch = EXTRACT(EPOCH FROM NOW())::BIGINT;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    tbl TEXT;
    trigger_name TEXT;
BEGIN
    FOREACH tbl IN ARRAY ARRAY[
        'users',
        'companies',
        'banks',
        'invoice_settings',
        'invoices',
        'categories',
        'products',
        'customers',
        'customer_products',
        'invoice_products'
    ]
    LOOP
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns c
            WHERE c.table_schema = 'public'
              AND c.table_name = tbl
              AND c.column_name = 'updated_at'
        )
        AND EXISTS (
            SELECT 1
            FROM information_schema.columns c
            WHERE c.table_schema = 'public'
              AND c.table_name = tbl
              AND c.column_name = 'updated_at_epoch'
        ) THEN
            trigger_name := 'trg_' || tbl || '_set_updated_epoch';

            IF NOT EXISTS (
                SELECT 1
                FROM pg_trigger t
                JOIN pg_class cl ON cl.oid = t.tgrelid
                JOIN pg_namespace ns ON ns.oid = cl.relnamespace
                WHERE t.tgname = trigger_name
                  AND ns.nspname = 'public'
                  AND cl.relname = tbl
            ) THEN
                EXECUTE format(
                    'CREATE TRIGGER %I BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION set_updated_epoch_fields()',
                    trigger_name,
                    tbl
                );
            END IF;
        END IF;
    END LOOP;
END;
$$;
