-- Migration 038: Add invoice-level overall discount
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS overall_discount NUMERIC(12, 2) DEFAULT 0;
