-- Migration: Add client_id support to devices table for multi-tenancy.

-- 1. Add client_id column, defaulting to 'bank_abc' for existing dev rows
ALTER TABLE devices ADD COLUMN IF NOT EXISTS client_id VARCHAR DEFAULT 'bank_abc';

-- 2. Drop the overly restrictive constraint we added in 002
ALTER TABLE devices DROP CONSTRAINT IF EXISTS unique_user_ref;

-- 3. Add the proper multi-tenant uniqueness constraint
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'unique_client_user'
          AND conrelid = 'devices'::regclass
    ) THEN
        ALTER TABLE devices ADD CONSTRAINT unique_client_user UNIQUE (client_id, user_ref);
    END IF;
END $$;

-- 4. Add index for faster query lookup
CREATE INDEX IF NOT EXISTS idx_devices_client_user ON devices(client_id, user_ref);
