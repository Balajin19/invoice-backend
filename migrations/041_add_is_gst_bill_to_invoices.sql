-- Migration 041: Add is_gst_bill column to invoices table

ALTER TABLE invoices
ADD COLUMN is_gst_bill BOOLEAN DEFAULT TRUE NOT NULL;

-- Create index for faster queries filtering by is_gst_bill
CREATE INDEX IF NOT EXISTS idx_invoices_is_gst_bill
ON invoices(is_gst_bill);
