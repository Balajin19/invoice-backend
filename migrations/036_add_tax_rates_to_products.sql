-- Migration 036: Support per-product tax rates (CGST, SGST, IGST)
-- This allows storing tax rates at the product level for invoice line items

-- Add tax rate columns to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS cgst_rate NUMERIC(5, 2) DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS sgst_rate NUMERIC(5, 2) DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS igst_rate NUMERIC(5, 2) DEFAULT 0;

-- Add tax rate columns to invoice_products table to store the rates at time of invoice
ALTER TABLE invoice_products ADD COLUMN IF NOT EXISTS cgst_rate NUMERIC(5, 2) DEFAULT 0;
ALTER TABLE invoice_products ADD COLUMN IF NOT EXISTS sgst_rate NUMERIC(5, 2) DEFAULT 0;
ALTER TABLE invoice_products ADD COLUMN IF NOT EXISTS igst_rate NUMERIC(5, 2) DEFAULT 0;

-- Add tax rate columns to customer_products table
ALTER TABLE customer_products ADD COLUMN IF NOT EXISTS cgst_rate NUMERIC(5, 2) DEFAULT 0;
ALTER TABLE customer_products ADD COLUMN IF NOT EXISTS sgst_rate NUMERIC(5, 2) DEFAULT 0;
ALTER TABLE customer_products ADD COLUMN IF NOT EXISTS igst_rate NUMERIC(5, 2) DEFAULT 0;
