-- Add targeting columns to zones table
ALTER TABLE zones ADD COLUMN IF NOT EXISTS countries text[];
ALTER TABLE zones ADD COLUMN IF NOT EXISTS device_types text[];
ALTER TABLE zones ADD COLUMN IF NOT EXISTS os text[];
ALTER TABLE zones ADD COLUMN IF NOT EXISTS browsers text[];
ALTER TABLE zones ADD COLUMN IF NOT EXISTS carriers text[];
ALTER TABLE zones ADD COLUMN IF NOT EXISTS connection_types text[];
