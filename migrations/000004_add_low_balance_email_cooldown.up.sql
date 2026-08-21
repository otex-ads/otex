-- Add column to track last low balance email sent timestamp
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS last_low_balance_email_at TIMESTAMPTZ;
