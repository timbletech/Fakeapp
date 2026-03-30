# Deployment Guide — Timble Auth API

## Architecture Overview

```mermaid
graph TD
    subgraph Bank[Bank - Client]
        BankApp["Bank Mobile App\n(End User Device)"]
        BankServer["Bank Backend Server"]
    end

    subgraph Timble[Timble - Linux Server 3.108.166.176]
        Nginx["Nginx\n(TLS termination :443)"]
        API["Timble API\n(Go binary :8097)"]
        PG["PostgreSQL\n(timble DB)"]
        Demo["Demo UI\n/demo/"]
    end

    BankApp -->|HTTPS API calls| Nginx
    BankServer -->|HTTPS API calls| Nginx
    Nginx -->|HTTP proxy| API
    API -->|SQL| PG
    API --> Demo
```

---

## Flow 1 — Device Authentication

```mermaid
sequenceDiagram
    participant BankApp as Bank Mobile App
    participant BankSrv as Bank Backend
    participant Timble as Timble API
    participant DB as PostgreSQL (Timble)

    Note over BankApp,Timble: One-time Device Registration
    BankApp->>Timble: POST /v1/device/register\n{client_id, user_ref, public_key, device_info}
    Timble->>DB: Store device binding + public key
    Timble-->>BankApp: {device_binding_id}

    Note over BankApp,Timble: Auth Flow (every login)
    BankApp->>Timble: POST /v1/auth/start\n{mode: "device", device_binding_id}
    Timble->>DB: Create auth session + challenge
    Timble-->>BankApp: {auth_session_id, challenge, challenge_id}

    BankApp->>BankApp: Sign challenge with EC private key

    BankApp->>Timble: POST /v1/auth/complete\n{auth_session_id, challenge_id, device_signature}
    Timble->>DB: Verify signature against stored public key
    Timble-->>BankApp: {decision: ALLOW, auth_context_token}

    Note over BankSrv,Timble: Bank backend validates token server-side
    BankSrv->>Timble: POST /v1/auth/verify\n{client_id, auth_context_token}
    Timble-->>BankSrv: {valid: true}
```

---

## Flow 2 — SIM Authentication

```mermaid
sequenceDiagram
    participant BankApp as Bank Mobile App
    participant BankSrv as Bank Backend
    participant Timble as Timble API
    participant Sekura as Sekura/XConnect (external)

    BankSrv->>Timble: POST /v1/sim/start\n{client_id, user_ref, msisdn}\nX-API-Key: ...
    Timble->>Sekura: Create SIM session
    Sekura-->>Timble: {session_uri}
    Timble-->>BankSrv: {auth_session_id, session_uri}

    BankSrv-->>BankApp: Pass session_uri to app
    BankApp->>Sekura: Open session_uri in device browser
    Sekura->>Sekura: SIM challenge resolved via mobile network

    loop Poll until final
        BankSrv->>Timble: POST /v1/sim/complete\n{auth_session_id}\nX-API-Key: ...
        Timble->>Sekura: Check session status
        Sekura-->>Timble: status
        Timble-->>BankSrv: 202 PENDING or 200 ALLOW/DENY
    end
```

---

## Flow 3 — Hybrid Authentication

```mermaid
sequenceDiagram
    participant BankApp as Bank Mobile App
    participant BankSrv as Bank Backend
    participant Timble as Timble API
    participant DB as PostgreSQL (Timble)
    participant Sekura as Sekura/XConnect (external)

    Note over BankApp,Timble: One-time device registration (same as Device Flow)

    BankSrv->>Timble: POST /v1/auth/start\n{mode: "hybrid", device_binding_id, msisdn}
    Timble->>DB: Create session + challenge
    Timble->>Sekura: Create SIM session
    Timble-->>BankSrv: {auth_session_id, challenge, sim.session_uri}

    BankSrv-->>BankApp: Forward challenge + session_uri

    par Bank App signs device challenge
        BankApp->>BankApp: Sign challenge with EC private key
    and Bank App triggers SIM check
        BankApp->>Sekura: Open session_uri in device browser
        Sekura->>Sekura: SIM resolved via mobile network
    end

    BankSrv->>Timble: POST /v1/auth/complete\n{auth_session_id, device_signature}
    Timble->>DB: Verify device signature
    Timble->>Sekura: Check SIM result

    alt Both factors verified
        Timble-->>BankSrv: {decision: ALLOW, auth_context_token}
    else SIM still in progress
        Timble-->>BankSrv: 202 {decision: PENDING} → poll again
    end
```

---

## Deployment Architecture

```mermaid
graph LR
    subgraph Bank[Bank - Client]
        BankApp["Bank Mobile App"]
        BankSrv["Bank Backend Server"]
    end

    subgraph Timble[Timble Linux Server - 3.108.166.176]
        Nginx["Nginx\n:443 / :80"]
        Go["Go API\n:8097"]
        PG["PostgreSQL\n:5432"]
        FS["Static Files\n/opt/timble/static/"]
        Migrations["Migrations\n/opt/timble/migrations/"]
    end

    BankApp -->|HTTPS| Nginx
    BankSrv -->|HTTPS| Nginx
    Nginx -->|proxy_pass| Go
    Go -->|DB_DSN| PG
    Go --> FS
    Go -->|on startup| Migrations
```

---

## Stack

| Component | Detail |
|-----------|--------|
| Language | Go 1.22.2 |
| Database | PostgreSQL |
| Server | Plain HTTP (`net/http`), reverse-proxied behind Nginx |
| Static UI | Served by Go at `/demo/` |

---

## Directory Layout (on server)

```
/opt/timble/
├── bin/
│   └── api              # compiled binary
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_user_ref_unique.sql
│   └── 003_add_client_id.sql
├── static/
│   ├── index.html
│   ├── app.js
│   └── style.css
└── .env
```

> The binary reads `migrations/` and `static/` **relative to its working directory**.
> Always run the binary from `/opt/timble/`.

---

## Step 1 — Build Binary

Run on your build machine (or the server if Go is installed):

```bash
cd /home/krishnaltp047/git_data/device_only

# Linux AMD64 (production server)
GOOS=linux GOARCH=amd64 go build -o bin/api ./cmd/api/main.go
```

---

## Step 2 — Copy Artifacts to Server

```bash
SERVER=user@3.108.166.176

ssh $SERVER "mkdir -p /opt/timble/bin /opt/timble/migrations /opt/timble/static"

scp bin/api                     $SERVER:/opt/timble/bin/api
scp migrations/*.sql            $SERVER:/opt/timble/migrations/
scp static/index.html \
    static/app.js \
    static/style.css            $SERVER:/opt/timble/static/

chmod +x /opt/timble/bin/api
```

---

## Step 3 — PostgreSQL Setup

```sql
CREATE DATABASE timble;
CREATE USER timble_user WITH PASSWORD 'strong_password';
GRANT ALL PRIVILEGES ON DATABASE timble TO timble_user;
```

> Migrations (`001`, `002`, `003`) are applied automatically on every startup.

---

## Step 4 — Environment File

Create `/opt/timble/.env`:

```env
# Server
SERVER_PORT=8097
SERVER_HOST=0.0.0.0
SERVER_DOMAIN=http://3.108.166.176:8097

# Database
DB_DSN=postgres://timble_user:strong_password@localhost:5432/timble?sslmode=disable

# Device Auth
DEV_MODE=false
DEV_PRIVATE_KEY=
CHALLENGE_EXPIRY_SECONDS=120
AUTH_TOKEN_EXPIRY_SECONDS=300

# Timble API (SIM)
TIMBLE_API_KEY=your_timble_api_key

# Sekura / XConnect
SEKURA_BASE_URL=https://in.safr.xconnect.net
SEKURA_CLIENT_KEY=your_sekura_client_key
SEKURA_CLIENT_SECRET=your_sekura_client_secret
SEKURA_REFRESH_TOKEN=your_sekura_refresh_token

# Session
SESSION_TTL_SECONDS=300
SESSION_MAX_ATTEMPTS=3
POLLING_MAX_RETRIES=3
POLLING_RETRY_DELAY_MS=2000
```

```bash
chmod 600 /opt/timble/.env
```

---

## Step 5 — Systemd Service

Create `/etc/systemd/system/timble.service`:

```ini
[Unit]
Description=Timble Auth API
After=network.target postgresql.service

[Service]
Type=simple
User=timble
WorkingDirectory=/opt/timble
EnvironmentFile=/opt/timble/.env
ExecStart=/opt/timble/bin/api
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

```bash
# Create system user (no login shell)
sudo useradd -r -s /sbin/nologin timble
sudo chown -R timble:timble /opt/timble

sudo systemctl daemon-reload
sudo systemctl enable timble
sudo systemctl start timble

# Verify
sudo systemctl status timble
sudo journalctl -u timble -f
```

---

## Step 6 — Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name yourdomain.com;

    ssl_certificate     /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass         http://127.0.0.1:8097;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
    }
}
```

```bash
# Free TLS via Let's Encrypt
sudo certbot --nginx -d yourdomain.com

sudo nginx -t
sudo systemctl reload nginx
```

> Update `SERVER_DOMAIN` in `.env` to `https://yourdomain.com` after enabling TLS,
> then `sudo systemctl restart timble`.

---

## Step 7 — Smoke Test

```bash
# Health / reachability
curl http://3.108.166.176:8097/health

# Demo UI
open http://3.108.166.176:8097/demo/

# Register a device
curl -X POST http://3.108.166.176:8097/v1/device/register \
  -H "Content-Type: application/json" \
  -d '{"client_id":"client_123","user_ref":"user_001","public_key":"<base64_pubkey>"}'
```

---

## Updating (Redeploy)

```bash
# 1. Build new binary
GOOS=linux GOARCH=amd64 go build -o bin/api ./cmd/api/main.go

# 2. Copy to server
scp bin/api user@3.108.166.176:/opt/timble/bin/api

# 3. Restart service
ssh user@3.108.166.176 "sudo systemctl restart timble"
```

---

## Routes Reference

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/device/register` | None | Register a device |
| POST | `/v1/device/revoke` | None | Revoke a device binding |
| PUT | `/v1/device/update` | None | Update device public key |
| GET | `/v1/device/check` | None | Check if device is active |
| POST | `/v1/auth/start` | None | Start auth session (device/sim/hybrid) |
| POST | `/v1/auth/complete` | None | Complete / poll auth session |
| POST | `/v1/auth/verify` | None | Verify an auth context token |
| POST | `/v1/hybrid/start` | None | Alias for auth/start forced to hybrid |
| POST | `/v1/hybrid/complete` | None | Alias for auth/complete |
| POST | `/v1/sim/start` | `X-API-Key` | Start SIM-only auth session |
| POST | `/v1/sim/complete` | `X-API-Key` | Poll SIM-only auth result |
| GET | `/v1/sim/redirect/{id}` | None | SIM redirect callback (browser) |
| GET | `/demo/` | None | Demo UI |

---

## API Integration Guide

### Base URL

```
http://YOUR_SERVER:8097
```

### Authentication

Only `/v1/sim/*` routes require an API key:

```
X-API-Key: your_timble_api_key
```

All other routes require no authentication header.

---

### Flow 1 — Device Authentication

Use this when the user's device holds a registered EC key pair.

#### Step 1: Register Device

```http
POST /v1/device/register
Content-Type: application/json

{
  "client_id": "client_123",
  "user_ref": "user_001",
  "public_key": "<base64-encoded EC public key>",
  "device_info": {
    "device_id": "unique-device-uuid",
    "platform": "android",
    "app_version": "1.0.0",
    "device_model": "Pixel 7",
    "os_version": "14"
  }
}
```

**Response `200`:**
```json
{
  "request_id": "req_abc123",
  "timestamp": "2026-03-17T10:00:00Z",
  "client_id": "client_123",
  "device_binding_id": "binding_xyz",
  "status": "registered"
}
```

> Save `device_binding_id` — required for every auth start.

---

#### Step 2: Start Auth Session

```http
POST /v1/auth/start
Content-Type: application/json

{
  "client_id": "client_123",
  "user_ref": "user_001",
  "mode": "device",
  "device_binding_id": "binding_xyz"
}
```

**Response `200`:**
```json
{
  "auth_session_id": "sess_abc",
  "mode": "device",
  "next_step": "SIGN_CHALLENGE",
  "device": {
    "challenge_id": "ch_xyz",
    "challenge": "<base64 bytes to sign>",
    "expires_in_seconds": 120
  },
  "status": "PENDING"
}
```

---

#### Step 3: Sign & Complete

Sign `device.challenge` with the device's EC private key, then:

```http
POST /v1/auth/complete
Content-Type: application/json

{
  "auth_session_id": "sess_abc",
  "mode": "device",
  "challenge_id": "ch_xyz",
  "device_signature": "<base64 EC signature>"
}
```

**Response `200` (success):**
```json
{
  "decision": "ALLOW",
  "auth_context_token": "<opaque token>",
  "expires_in_seconds": 300,
  "status": "SUCCESS"
}
```

---

#### Step 4: Verify Token (optional — server-side validation)

```http
POST /v1/auth/verify
Content-Type: application/json

{
  "client_id": "client_123",
  "auth_context_token": "<opaque token>"
}
```

**Response `200`:**
```json
{
  "valid": true,
  "expires_in_seconds": 240,
  "status": "valid"
}
```

---

### Flow 2 — SIM Authentication

Use this when you want to verify the SIM card present in the device.

> Requires `X-API-Key` header on all `/v1/sim/*` calls.

#### Step 1: Start SIM Session

```http
POST /v1/sim/start
Content-Type: application/json
X-API-Key: your_timble_api_key

{
  "client_id": "client_123",
  "user_ref": "user_001",
  "msisdn": "917905968734"
}
```

**Response `200`:**
```json
{
  "auth_session_id": "sim_sess_abc",
  "session_uri": "https://...",
  "expires_in": 300,
  "next_step": "REDIRECT_USER",
  "instructions": "Redirect user to session_uri"
}
```

> Redirect the user's device browser to `session_uri` to trigger the SIM check.

---

#### Step 2: Poll for Result

Call repeatedly until `decision` is `ALLOW` or `DENY` (not `PENDING`):

```http
POST /v1/sim/complete
Content-Type: application/json
X-API-Key: your_timble_api_key

{
  "auth_session_id": "sim_sess_abc"
}
```

**Response `202` (still pending):**
```json
{
  "auth_session_id": "sim_sess_abc",
  "status": "PENDING",
  "attempts_remaining": 2
}
```

**Response `200` (final):**
```json
{
  "decision": "ALLOW",
  "reason_code": "SIM_MATCH",
  "sim_swap_safe": true,
  "device_match": true,
  "completed_at": "2026-03-17T10:01:00Z"
}
```

---

### Flow 3 — Hybrid Authentication

Combines SIM verification **and** device cryptography in one session.

#### Step 1: Register Device (same as Device Flow — one-time)

#### Step 2: Start Hybrid Session

```http
POST /v1/auth/start
Content-Type: application/json

{
  "client_id": "client_123",
  "user_ref": "user_001",
  "mode": "hybrid",
  "msisdn": "917905968734",
  "device_binding_id": "binding_xyz",
  "device_info": {
    "device_id": "unique-device-uuid",
    "platform": "android"
  }
}
```

**Response `200`:**
```json
{
  "auth_session_id": "sess_hyb",
  "mode": "hybrid",
  "next_step": "SIGN_CHALLENGE_AND_REDIRECT_SIM",
  "device": {
    "challenge_id": "ch_xyz",
    "challenge": "<base64 bytes to sign>",
    "expires_in_seconds": 120
  },
  "sim": {
    "auth_session_id": "sim_sess_hyb",
    "session_uri": "https://...",
    "expires_in_seconds": 300
  },
  "status": "PENDING"
}
```

> Redirect the user to `sim.session_uri` **and** sign `device.challenge` in parallel.

---

#### Step 3: Complete Hybrid Auth

```http
POST /v1/auth/complete
Content-Type: application/json

{
  "auth_session_id": "sess_hyb",
  "mode": "hybrid",
  "challenge_id": "ch_xyz",
  "device_signature": "<base64 EC signature>"
}
```

**Response `200` (both factors done):**
```json
{
  "decision": "ALLOW",
  "auth_context_token": "<opaque token>",
  "expires_in_seconds": 300,
  "status": "SUCCESS"
}
```

**Response `202` (SIM still pending):**
```json
{
  "decision": "PENDING",
  "reason_code": "SIM_PENDING",
  "next_step": "SIM_CHALLENGE_REQUIRED",
  "attempts_remaining": 2,
  "status": "PENDING"
}
```

> Poll `/v1/auth/complete` again with the same `auth_session_id` until final.

---

### Device Management

#### Check Device Status

```http
GET /v1/device/check?client_id=client_123&user_ref=user_001
```

**Response `200`:**
```json
{
  "has_active_device": true,
  "device_binding_id": "binding_xyz",
  "status": "active"
}
```

---

#### Update Device Key (key rotation)

```http
PUT /v1/device/update
Content-Type: application/json

{
  "client_id": "client_123",
  "user_ref": "user_001",
  "public_key": "<new base64 EC public key>",
  "device_info": {
    "device_id": "unique-device-uuid",
    "platform": "android"
  }
}
```

---

#### Revoke Device

```http
POST /v1/device/revoke
Content-Type: application/json

{
  "client_id": "client_123",
  "user_ref": "user_001",
  "device_binding_id": "binding_xyz"
}
```

**Response `200`:**
```json
{
  "status": "revoked"
}
```

---

### Error Response Format

All errors follow the same envelope:

```json
{
  "error": "error_code",
  "message": "Human readable description",
  "request_id": "req_abc123"
}
```

| HTTP Status | `error` value | Meaning |
|-------------|---------------|---------|
| 400 | `invalid_request` | Missing or malformed fields |
| 400 | `validation_error` | Field value failed validation |
| 401 | `unauthorized` | Missing or invalid `X-API-Key` |
| 500 | `internal_error` | Server-side failure |
| 502 | `upstream_error` | Sekura / XConnect upstream failure |
