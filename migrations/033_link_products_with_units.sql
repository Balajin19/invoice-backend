ALTER TABLE products
    ADD COLUMN IF NOT EXISTS unit_id UUID;

INSERT INTO units (unit_name)
SELECT DISTINCT TRIM(p.unit)
FROM products p
WHERE TRIM(COALESCE(p.unit, '')) <> ''
  AND NOT EXISTS (
      SELECT 1
      FROM units u
      WHERE LOWER(TRIM(u.unit_name)) = LOWER(TRIM(p.unit))
  );

UPDATE products p
SET unit_id = u.unit_id
FROM units u
WHERE p.unit_id IS NULL
  AND LOWER(TRIM(p.unit)) = LOWER(TRIM(u.unit_name));

ALTER TABLE products
    ALTER COLUMN unit_id SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_products_unit_id'
    ) THEN
        ALTER TABLE products
            ADD CONSTRAINT fk_products_unit_id
            FOREIGN KEY (unit_id) REFERENCES units(unit_id) ON DELETE RESTRICT;
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_products_unit_id ON products(unit_id);

DROP INDEX IF EXISTS idx_products_name_unit_category_unique;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM (
            SELECT LOWER(TRIM(product_name)) AS product_name_key,
                   unit_id,
                   category_id,
                   COUNT(*) AS cnt
            FROM products
            GROUP BY LOWER(TRIM(product_name)), unit_id, category_id
            HAVING COUNT(*) > 1
        ) duplicates
    ) THEN
        CREATE UNIQUE INDEX IF NOT EXISTS idx_products_name_unit_category_unique
            ON products (LOWER(TRIM(product_name)), unit_id, category_id);
    ELSE
        RAISE NOTICE 'Skipping idx_products_name_unit_category_unique creation because duplicate product records exist after unit mapping';
    END IF;
END $$;
