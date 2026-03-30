# Timble API — Integration Guide

Complete reference for integrating the **Bank backend** and **Bank mobile app** with the Timble authentication API.

---

## Base URL & Headers

```
Base URL:  http://3.108.166.176:8097
```

| Header           | Required on        | Value                |
| ---------------- | ------------------ | -------------------- |
| `Content-Type` | All POST/PUT       | `application/json` |
| `X-API-Key`    | `/v1/sim/*` only | Your Timble API key  |

---

## Common Response Envelope

Every response — success or error — includes:

```json
{
  "request_id": "req_<uuid>",
  "timestamp": "2026-03-17T10:00:00Z",
  ...
}
```

Every error response:

```json
{
  "request_id": "req_abc",
  "error": "error_code",
  "message": "Human readable description"
}
```

| HTTP Status | `error`            | Meaning                            |
| ----------- | -------------------- | ---------------------------------- |
| `400`     | `invalid_request`  | Missing or malformed fields        |
| `400`     | `validation_error` | Field value failed validation      |
| `401`     | `unauthorized`     | Missing or invalid `X-API-Key`   |
| `500`     | `internal_error`   | Server-side failure                |

---

## Device Management

### Register Device

> Called once when a user installs / first opens the bank app.

```http
POST /v1/device/register
Content-Type: application/json
```

**Request**

```json
{
  "client_id": "client_123",
  "user_ref": "user_001",
  "public_key": "<base64-encoded EC public key>",
  "device_info": {
    "device_id": "unique-device-uuid",
    "platform": "android",
    "app_version": "1.0.0",
    "device_model": "Pixel 7",
    "os_version": "14",
    "ip_address": "192.168.1.10"
  }
}
```

**Response `200`**

```json
{
  "request_id": "req_abc123",
  "timestamp": "2026-03-17T10:00:00Z",
  "client_id": "client_123",
  "device_binding_id": "binding_xyz",
  "status": "registered"
}
```

> **Save `device_binding_id`** locally — required for every subsequent auth call.

---

### Check Device Status

```http
GET /v1/device/check?client_id=client_123&user_ref=user_001
```

**Response `200`**

```json
{
  "request_id": "req_abc",
  "timestamp": "2026-03-17T10:00:00Z",
  "has_active_device": true,
  "device_binding_id": "binding_xyz",
  "status": "active"
}
```

---

### Update Device Key

> Use for key rotation or app reinstall.

```http
PUT /v1/device/update
Content-Type: application/json
```

**Request**

```json
{
  "client_id": "client_123",
  "user_ref": "user_001",
  "public_key": "<new base64 EC public key>",
  "device_info": {
    "device_id": "unique-device-uuid",
    "platform": "android",
    "app_version": "1.1.0"
  }
}
```

**Response `200`**

```json
{
  "request_id": "req_abc",
  "timestamp": "2026-03-17T10:00:00Z",
  "client_id": "client_123",
  "device_binding_id": "binding_xyz",
  "status": "updated"
}
```

---

### Revoke Device

```http
POST /v1/device/revoke
Content-Type: application/json
```

**Request**

```json
{
  "client_id": "client_123",
  "user_ref": "user_001",
  "device_binding_id": "binding_xyz"
}
```

**Response `200`**

```json
{
  "request_id": "req_abc",
  "timestamp": "2026-03-17T10:00:00Z",
  "status": "revoked"
}
```

---

## Auth Flows

### Flow 1 — Device Authentication

**Step 1: Start session**

```http
POST /v1/auth/start
Content-Type: application/json
```

```json
{
  "client_id": "client_123",
  "user_ref": "user_001",
  "mode": "device",
  "device_binding_id": "binding_xyz"
}
```

**Response `200`**

```json
{
  "request_id": "req_abc",
  "timestamp": "2026-03-17T10:00:00Z",
  "client_id": "client_123",
  "mode": "device",
  "auth_session_id": "sess_abc",
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

**Step 2: Sign challenge on device, then complete**

```http
POST /v1/auth/complete
Content-Type: application/json
```

```json
{
  "auth_session_id": "sess_abc",
  "mode": "device",
  "challenge_id": "ch_xyz",
  "device_signature": "<base64 EC signature of challenge bytes>"
}
```

**Response `200` — success**

```json
{
  "request_id": "req_abc",
  "timestamp": "2026-03-17T10:00:00Z",
  "client_id": "client_123",
  "mode": "device",
  "auth_session_id": "sess_abc",
  "decision": "ALLOW",
  "auth_context_token": "<opaque token>",
  "expires_in_seconds": 300,
  "status": "SUCCESS"
}
```

**Response `200` — denied**

```json
{
  "decision": "DENY",
  "reason_code": "SIGNATURE_INVALID",
  "reason_message": "Device signature verification failed",
  "status": "FAILED"
}
```

---

**Step 3: Verify token (bank backend)**

```http
POST /v1/auth/verify
Content-Type: application/json
```

```json
{
  "client_id": "client_123",
  "auth_context_token": "<opaque token>"
}
```

**Response `200`**

```json
{
  "request_id": "req_abc",
  "timestamp": "2026-03-17T10:00:00Z",
  "client_id": "client_123",
  "valid": true,
  "expires_in_seconds": 240,
  "status": "valid"
}
```

---

### Flow 2 — SIM Authentication

> All `/v1/sim/*` calls require `X-API-Key` header.
> The bank **backend** makes these calls — not the mobile app directly.

**Step 1: Start SIM session**

```http
POST /v1/sim/start
Content-Type: application/json
X-API-Key: your_timble_api_key
```

```json
{
  "client_id": "client_123",
  "user_ref": "user_001",
  "msisdn": "917905968734"
}
```

**Response `200`**

```json
{
  "request_id": "req_abc",
  "auth_session_id": "sim_sess_abc",
  "session_uri": "http://3.108.166.176:8097/v1/sim/redirect/sess_3644c323-30a0-4399-bf52-1455846016e3",
  "expires_in": 300,
  "next_step": "REDIRECT_USER",
  "instructions": "Redirect user device browser to session_uri",
  "sim_swap_check": {
    "result": false,
    "date": "2026-03-10T00:00:00Z",
    "seconds": 604800
  },
  "operator_lookup": {
    "regionCode": "IN",
    "operatorName": "Jio",
    "mcc": "404",
    "mnc": "50"
  }
}
```

> Redirect the user's mobile app browser to `session_uri`. resolves the SIM challenge silently over the mobile network.

---

**Step 2: Poll for result**

Poll until `decision` is `ALLOW` or `DENY`:

```http
POST /v1/sim/complete
Content-Type: application/json
X-API-Key: your_timble_api_key
```

```json
{
  "auth_session_id": "sim_sess_abc"
}
```

**Response `202` — still pending**

```json
{
  "request_id": "req_abc",
  "auth_session_id": "sim_sess_abc",
  "status": "PENDING",
  "message": "SIM challenge not yet resolved",
  "attempts_remaining": 2
}
```

**Response `200` — final decision**

```json
{
  "request_id": "req_abc",
  "auth_session_id": "sim_sess_abc",
  "decision": "ALLOW",
  "reason_code": "SIM_MATCH",
  "reason_message": "SIM verified successfully",
  "device_match": true,
  "sim_swap_safe": true,
  "completed_at": "2026-03-17T10:01:00Z"
}
```

---

### Flow 3 — Hybrid Authentication

Combines device cryptography **and** SIM verification in one session.

**Step 1: Start hybrid session (bank backend)**

```http
POST /v1/auth/start
Content-Type: application/json
```

```json
{
  "client_id": "client_123",
  "user_ref": "user_001",
  "mode": "hybrid",
  "device_binding_id": "binding_xyz",
  "msisdn": "917905968734",
  "device_info": {
    "device_id": "unique-device-uuid",
    "platform": "android"
  }
}
```

**Response `200`**

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
    "session_uri": "http://3.108.166.176:8097/v1/sim/redirectsess_3644c323-30a0-4399-bf52-1455846016e3",
    "expires_in_seconds": 300
  },
  "status": "PENDING"
}
```

> Send both `device.challenge` and `sim.session_uri` to the mobile app.
> The app must **sign the challenge** and **open session_uri** in parallel.

---

**Step 2: Complete hybrid (submit device signature)**

```http
POST /v1/auth/complete
Content-Type: application/json
```

```json
{
  "auth_session_id": "sess_hyb",
  "mode": "hybrid",
  "challenge_id": "ch_xyz",
  "device_signature": "<base64 EC signature>"
}
```

**Response `200` — both factors done**

```json
{
  "decision": "ALLOW",
  "auth_context_token": "<opaque token>",
  "expires_in_seconds": 300,
  "status": "SUCCESS"
}
```

**Response `202` — SIM still pending**

```json
{
  "decision": "PENDING",
  "reason_code": "SIM_PENDING",
  "next_step": "SIM_CHALLENGE_REQUIRED",
  "attempts_remaining": 2,
  "status": "PENDING"
}
```

> Poll `/v1/auth/complete` again with the same `auth_session_id` (no signature needed on retry) until final.

---

## Quick Reference — All Endpoints

| Method   | Path                      | Auth          | Who calls it  | Description                            |
| -------- | ------------------------- | ------------- | ------------- | -------------------------------------- |
| `POST` | `/v1/device/register`   | None          | Mobile App    | Register device + public key           |
| `GET`  | `/v1/device/check`      | None          | Backend / App | Check if device is active              |
| `PUT`  | `/v1/device/update`     | None          | Mobile App    | Rotate device public key               |
| `POST` | `/v1/device/revoke`     | None          | Backend       | Revoke device binding                  |
| `POST` | `/v1/auth/start`        | None          | Backend       | Start auth session (device/sim/hybrid) |
| `POST` | `/v1/auth/complete`     | None          | Backend       | Complete / poll auth session           |
| `POST` | `/v1/auth/verify`       | None          | Backend       | Verify auth context token              |
| `POST` | `/v1/hybrid/start`      | None          | Backend       | Alias for auth/start forced to hybrid  |
| `POST` | `/v1/hybrid/complete`   | None          | Backend       | Alias for auth/complete                |
| `POST` | `/v1/sim/start`         | `X-API-Key` | Backend       | Start SIM-only session                 |
| `POST` | `/v1/sim/complete`      | `X-API-Key` | Backend       | Poll SIM-only result                   |
| `GET`  | `/v1/sim/redirect/{id}` | None          | Browser (App) | SIM redirect callback                  |

---

## curl Examples

### Register a device

```bash
curl -X POST http://3.108.166.176:8097/v1/device/register \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client_123",
    "user_ref": "user_001",
    "public_key": "BASE64_PUBLIC_KEY",
    "device_info": {
      "device_id": "device-uuid-001",
      "platform": "android",
      "app_version": "1.0.0"
    }
  }'
```

### Start device auth

```bash
curl -X POST http://3.108.166.176:8097/v1/auth/start \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client_123",
    "user_ref": "user_001",
    "mode": "device",
    "device_binding_id": "binding_xyz"
  }'
```

### Complete device auth

```bash
curl -X POST http://3.108.166.176:8097/v1/auth/complete \
  -H "Content-Type: application/json" \
  -d '{
    "auth_session_id": "sess_abc",
    "mode": "device",
    "challenge_id": "ch_xyz",
    "device_signature": "BASE64_SIGNATURE"
  }'
```

### Start SIM auth

```bash
curl -X POST http://3.108.166.176:8097/v1/sim/start \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_timble_api_key" \
  -d '{
    "client_id": "client_123",
    "user_ref": "user_001",
    "msisdn": "917905968734"
  }'
```

### Poll SIM result

```bash
curl -X POST http://3.108.166.176:8097/v1/sim/complete \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_timble_api_key" \
  -d '{"auth_session_id": "sim_sess_abc"}'
```

### Verify token

```bash
curl -X POST http://3.108.166.176:8097/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client_123",
    "auth_context_token": "TOKEN"
  }'
```
