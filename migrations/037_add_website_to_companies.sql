-- Migration 037: Add company website information for PDF display
-- Website will be displayed in invoice PDF if available, otherwise only email is shown

ALTER TABLE companies ADD COLUMN IF NOT EXISTS website VARCHAR(255) DEFAULT '';
