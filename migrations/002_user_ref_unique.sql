-- Migration: Make user_ref unique in the devices table.

-- 1. Ensure any duplicate user_refs are either deleted or handled.
-- 2. Skip this migration once multi-tenant uniqueness (unique_client_user) exists.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'unique_client_user'
          AND conrelid = 'devices'::regclass
    ) THEN
        RETURN;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'unique_user_ref'
          AND conrelid = 'devices'::regclass
    ) THEN
        ALTER TABLE devices
        ADD CONSTRAINT unique_user_ref UNIQUE (user_ref);
    END IF;
END $$;
