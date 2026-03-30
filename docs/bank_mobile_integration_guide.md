# Bank Mobile Integration Guide

## 1) Authentication Modes

- `device`: Device cryptographic challenge-sign flow only.
- `sim`: SIM/XConnect verification only.
- `hybrid`: SIM verification + device cryptographic signature.

Mode can be sent in request payload (`mode`) and is enforced by server-side client policy.

## 2) Client-Level Mode Configuration

Set these environment variables on the API server:

```env
DEFAULT_AUTH_MODE=device
CLIENT_AUTH_MODES=client_123:hybrid,client_999:sim
```

Rules:

- If a client has an explicit override in `CLIENT_AUTH_MODES`, request `mode` must match it.
- If no override exists, request `mode` is accepted.
- If request `mode` is empty, server uses the configured default/override.

## 3) API Endpoints

Full schema: [openapi.yaml](/home/krishnaltp047/git_data/device_only/docs/openapi.yaml)

- `POST /v1/device/register`
- `POST /v1/auth/start`
- `POST /v1/auth/complete`
- `POST /v1/auth/verify`
- `POST /v1/hybrid/start` (alias)
- `POST /v1/hybrid/complete` (alias)
- `POST /v1/sim/start` and `POST /v1/sim/complete` (direct SIM legacy/internal)

## 4) Device-Only Flow

1. Register device with public key.
2. Call `/v1/auth/start` with `mode=device`.
3. Mobile app signs returned `device.challenge` with device private key.
4. Call `/v1/auth/complete` with `mode=device`, `auth_session_id`, `challenge_id`, `device_signature`.
5. Call `/v1/auth/verify` with `auth_context_token`.

### Start Request (device)

```json
{
  "client_id": "client_123",
  "user_ref": "user_457",
  "action": "login",
  "mode": "device",
  "device_binding_id": "bind_xxx",
  "device_info": {
    "device_id": "ios-device-001",
    "platform": "ios"
  }
}
```

## 5) SIM-Only Flow (XConnect)

1. Call `/v1/auth/start` with `mode=sim` and `msisdn`.
2. Open `sim.session_uri` on mobile data.
3. Poll `/v1/auth/complete` with `mode=sim` and top-level `auth_session_id`.
4. On `202` + `decision=PENDING`, retry.
5. On `ALLOW`, server returns `auth_context_token`.

### Start Request (sim)

```json
{
  "client_id": "client_123",
  "user_ref": "user_457",
  "action": "check",
  "mode": "sim",
  "msisdn": "917905968734"
}
```

## 6) Hybrid Flow (SIM + Device Signature)

1. Call `/v1/auth/start` with `mode=hybrid`, `msisdn`, and device binding details.
2. App opens `sim.session_uri` on mobile data and also signs `device.challenge`.
3. Call `/v1/auth/complete` with `mode=hybrid`, top-level `auth_session_id`, and `device_signature`.
4. If SIM is still pending, server responds `202` with `decision=PENDING`.
5. Once SIM is `ALLOW`, server verifies device signature and returns final ALLOW/DENY.

### Start Request (hybrid)

```json
{
  "client_id": "client_123",
  "user_ref": "user_457",
  "action": "login",
  "mode": "hybrid",
  "msisdn": "917905968734",
  "device_binding_id": "bind_xxx",
  "device_info": {
    "device_id": "ios-device-001",
    "platform": "ios"
  }
}
```

### Complete Request (hybrid)

```json
{
  "client_id": "client_123",
  "mode": "hybrid",
  "auth_session_id": "auth_xxx",
  "device_signature": "base64-asn1-signature"
}
```

## 7) Challenge-Sign Contract

- Challenge is base64 string from `device.challenge`.
- Signature must be ASN.1 ECDSA signature encoded in base64.
- Signature verification is server-side using registered device public key.
- In `DEV_MODE=true`, `demo_bypass_signature` is allowed for local UI/demo only.

## 8) SIM/XConnect Notes

- SIM flow is integrated through the SIM service/provider layer.
- Existing provider implementation uses Sekura/XConnect endpoints.
- Replace provider internals to switch to production telecom/XConnect tenancy without changing API handlers.

## 9) Response Handling Recommendations

- Treat `202` from `/v1/auth/complete` as retryable pending state.
- Treat `decision=DENY` as terminal.
- Persist `auth_context_token` only for short-lived server-side handoff; do not store long term on device.

## 10) Local Validation

Run:

```bash
./test_api.sh
```

The script covers:

- Device registration + device auth
- Unified sim mode start/complete
- Unified hybrid mode start/complete
- Direct legacy sim endpoints
