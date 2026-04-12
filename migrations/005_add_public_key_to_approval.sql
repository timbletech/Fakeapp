-- Add requesting_public_key column to device_approval_requests
ALTER TABLE device_approval_requests ADD COLUMN IF NOT EXISTS requesting_public_key TEXT;
