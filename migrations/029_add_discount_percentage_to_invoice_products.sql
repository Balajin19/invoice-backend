-- Increase account_number column size for encrypted values
ALTER TABLE banks
ALTER COLUMN account_number TYPE VARCHAR(256);
