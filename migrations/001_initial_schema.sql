CREATE TABLE IF NOT EXISTS devices (
    device_binding_id  VARCHAR      PRIMARY KEY,
    user_ref           VARCHAR      NOT NULL,
    public_key         TEXT         NOT NULL,
    device_id          VARCHAR,
    platform           VARCHAR,
    device_model       VARCHAR,
    os_version         VARCHAR,
    created_at         TIMESTAMP    DEFAULT NOW(),
    status             VARCHAR      DEFAULT 'ACTIVE'
);

CREATE INDEX IF NOT EXISTS idx_devices_user_ref ON devices(user_ref);

CREATE TABLE IF NOT EXISTS auth_sessions (
    auth_session_id    VARCHAR      PRIMARY KEY,
    user_ref           VARCHAR      NOT NULL,
    challenge          TEXT         NOT NULL,
    challenge_id       VARCHAR,
    device_binding_id  VARCHAR,
    expires_at         TIMESTAMP,
    status             VARCHAR      DEFAULT 'PENDING',
    created_at         TIMESTAMP    DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_ref ON auth_sessions(user_ref);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_device_binding_id ON auth_sessions(device_binding_id);

CREATE TABLE IF NOT EXISTS auth_context_tokens (
    token       VARCHAR      PRIMARY KEY,
    user_ref    VARCHAR      NOT NULL,
    expires_at  TIMESTAMP,
    status      VARCHAR      DEFAULT 'ACTIVE',
    created_at  TIMESTAMP    DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_context_tokens_user_ref ON auth_context_tokens(user_ref);

CREATE TABLE IF NOT EXISTS audit_logs (
    id          SERIAL       PRIMARY KEY,
    user_ref    VARCHAR,
    action      VARCHAR,
    decision    VARCHAR,
    ip_address  VARCHAR,
    device_id   VARCHAR,
    created_at  TIMESTAMP    DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_ref ON audit_logs(user_ref);
