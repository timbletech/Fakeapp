-- Migration: Add device_approval_requests table for new-device verification flow.
-- When a user logs in from an unknown device, an approval request is sent to
-- the user's main (trusted) device. This table tracks that approval lifecycle.

CREATE TABLE IF NOT EXISTS device_approval_requests (
    id                  VARCHAR      PRIMARY KEY,
    client_id           VARCHAR      NOT NULL,
    user_ref            VARCHAR      NOT NULL,
    requesting_device_id VARCHAR     NOT NULL,
    requesting_device_info JSONB,
    main_device_binding_id VARCHAR   NOT NULL REFERENCES devices(device_binding_id),
    status              VARCHAR      NOT NULL DEFAULT 'PENDING',  -- PENDING | APPROVED | DENIED | EXPIRED
    created_at          TIMESTAMP    DEFAULT NOW(),
    expires_at          TIMESTAMP    NOT NULL,
    resolved_at         TIMESTAMP,
    resolved_by         VARCHAR      -- device_binding_id of the device that approved/denied
);

CREATE INDEX IF NOT EXISTS idx_device_approval_user ON device_approval_requests(client_id, user_ref);
CREATE INDEX IF NOT EXISTS idx_device_approval_status ON device_approval_requests(status) WHERE status = 'PENDING';
