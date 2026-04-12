# Timble Mode B (Device + SIM + Hybrid Authentication)

This repository contains the backend implementation for Timble Mode B authentication using Go and PostgreSQL.

It supports:
- Device-only cryptographic challenge-response (ECDSA)
- SIM-only verification through XConnect/Sekura integration
- Hybrid mode (SIM verification + device challenge-sign)
- New device approval flow (unknown device verification via trusted device)

## Features

- **Blazing Fast API**: Uses Go 1.22 raw HTTP routing (`http.ServeMux`), zero external framework overhead.
- **Cryptographically Secure**: 32-byte randomized challenges signed with ECDSA (P-256).
- **PostgreSQL Database Storage**: Scalable schemas with indices covering user references and binding lookups. fully ready for UUID primary keys.
- **Client-Level Mode Policy**: Configure default and per-client auth mode (`device`, `sim`, `hybrid`).
- **New Device Verification**: When an unknown device attempts login, the system auto-triggers an approval request to the user's trusted device. Once approved, the new device is registered and can proceed with normal auth.

## API Documentation

- OpenAPI spec: [docs/openapi.yaml](docs/openapi.yaml)
- Bank mobile integration guide: [docs/bank_mobile_integration_guide.md](docs/bank_mobile_integration_guide.md)

## Architecture

```text
cmd/
  api/           # Main HTTP server
  devsigner/     # CLI for local simulation of device signing
internal/
  config/        # Environment logic
  crypto/        # ECDSA validation and generator utilities
  handlers/      # REST API endpoints layer
  models/        # Core domain entities
  orchestration/ # In-memory session store for multi-mode auth
  repository/    # DB queries
  service/       # Business rules (auth, device, device-verify)
  sim/           # SIM-only authentication (Sekura/XConnect)
migrations/      # .sql files (auto-applied on startup)
```

## Running the Server

1. Start PostgreSQL server on your machine.
2. Create the Database (but no need to run migrations, the server does that automatically!):
```bash
psql -U postgres -c "CREATE DATABASE timble;"
```
3. Prepare configuration:
```bash
cp .env.example .env
```
(Adjust DB credentials if necessary)
4. Start Server:
```bash
go run cmd/api/main.go
```

## Local Dev Simulation (DEV_MODE)

Mobile devices typically sign the base64 challenge securely in their Secure Enclaves or Android Keystores. For development, we provide a `devsigner` CLI.

Ensure `DEV_MODE=true` is set in `.env`.

To run an end-to-end integration test locally, ensure your database connection inside `.env` is correct and run:

```bash
chmod +x test_api.sh
./test_api.sh
```

**What the tester does:**
1. Generates an ECDSA KeyPair using `devsigner keygen`
2. Updates `.env` to hold the dev private key.
3. Registers the device via POST `/v1/device/register` sending the generated public key.
4. Starts auth session via POST `/v1/auth/start` to get a challenge.
5. Pushes the challenge locally to `devsigner sign --challenge <base64>` representing the mobile app's responsibility.
6. Returns `device_signature` to POST `/v1/auth/complete`.
7. Asserts context token against POST `/v1/auth/verify`.

## Curl Reference (All API Calls)

The following sequence contains all curl requests currently used in `test_api.sh`.

```bash
# Configuration
API_BASE="http://localhost:8097"
TIMBLE_API_KEY="timble-test-key-2026"
CLIENT_ID="client_123"
USER_REF="user_457"
DEVICE_ID="dev_789"
PUBLIC_KEY="LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFTnZMUXgyaXRReWdoV1RzWFgwajlLRzdvUE4zLwpybUZBNXBLY2h2VFQrU3BTWWhyQm5iMlo1clRwZ0lLVW9IWFUzTEM0aGhOK2NJaDIrKzBHaHdnSHlnPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="

# 1) Device register
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/v1/device/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"device_info\": {
      \"device_id\": \"$DEVICE_ID\",
      \"platform\": \"ios\",
      \"device_model\": \"iPhone 15\",
      \"os_version\": \"17.0\"
    },
    \"public_key\": \"$PUBLIC_KEY\"
  }")
echo "$REGISTER_RESPONSE" | jq .
BINDING_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.device_binding_id')

# 2) Device check
curl -s -G "$API_BASE/v1/device/check" \
  --data-urlencode "client_id=$CLIENT_ID" \
  --data-urlencode "user_ref=$USER_REF" \
  --data-urlencode "device_id=$DEVICE_ID" | jq .

# 3) Device update
curl -s -X PUT "$API_BASE/v1/device/update" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"device_info\": {
      \"device_id\": \"$DEVICE_ID\",
      \"platform\": \"ios\",
      \"os_version\": \"17.1\"
    },
    \"public_key\": \"$PUBLIC_KEY\"
  }" | jq .

# 4) Unified auth start (device mode)
AUTH_START_RESP=$(curl -s -X POST "$API_BASE/v1/auth/start" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"action\": \"login\",
    \"mode\": \"device\",
    \"device_binding_id\": \"$BINDING_ID\",
    \"device_info\": {
      \"device_id\": \"$DEVICE_ID\",
      \"platform\": \"ios\"
    }
  }")
echo "$AUTH_START_RESP" | jq .
SESSION_ID=$(echo "$AUTH_START_RESP" | jq -r '.auth_session_id')
CHALLENGE_ID=$(echo "$AUTH_START_RESP" | jq -r '.device.challenge_id')
CHALLENGE=$(echo "$AUTH_START_RESP" | jq -r '.device.challenge')

# 5) Sign challenge (dev tool)
SIGNATURE=$(go run cmd/devsigner/main.go sign --challenge "$CHALLENGE")

# 6) Unified auth complete (device mode)
AUTH_COMP_RESP=$(curl -s -X POST "$API_BASE/v1/auth/complete" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"mode\": \"device\",
    \"auth_session_id\": \"$SESSION_ID\",
    \"challenge_id\": \"$CHALLENGE_ID\",
    \"device_signature\": \"$SIGNATURE\"
  }")
echo "$AUTH_COMP_RESP" | jq .
CONTEXT_TOKEN=$(echo "$AUTH_COMP_RESP" | jq -r '.auth_context_token')

# 7) Verify token
curl -s -X POST "$API_BASE/v1/auth/verify" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"auth_context_token\": \"$CONTEXT_TOKEN\"
  }" | jq .

# 8) Unified auth start (sim mode)
SIM_UNIFIED_START_RESP=$(curl -s -X POST "$API_BASE/v1/auth/start" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"action\": \"check\",
    \"mode\": \"sim\",
    \"msisdn\": \"917905968734\"
  }")
echo "$SIM_UNIFIED_START_RESP" | jq .
SIM_UNIFIED_SESSION_ID=$(echo "$SIM_UNIFIED_START_RESP" | jq -r '.auth_session_id')

# 9) Unified auth complete (sim mode)
curl -s -X POST "$API_BASE/v1/auth/complete" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"mode\": \"sim\",
    \"auth_session_id\": \"$SIM_UNIFIED_SESSION_ID\"
  }" | jq .

# 10) Unified auth start (hybrid mode)
HYBRID_START_RESP=$(curl -s -X POST "$API_BASE/v1/auth/start" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"action\": \"login\",
    \"mode\": \"hybrid\",
    \"msisdn\": \"917905968734\",
    \"device_binding_id\": \"$BINDING_ID\",
    \"device_info\": {
      \"device_id\": \"$DEVICE_ID\",
      \"platform\": \"ios\"
    }
  }")
echo "$HYBRID_START_RESP" | jq .
HYBRID_SESSION_ID=$(echo "$HYBRID_START_RESP" | jq -r '.auth_session_id')
HYBRID_CHALLENGE=$(echo "$HYBRID_START_RESP" | jq -r '.device.challenge')

# 11) Unified auth complete (hybrid mode)
HYBRID_SIGNATURE=$(go run cmd/devsigner/main.go sign --challenge "$HYBRID_CHALLENGE")
curl -s -X POST "$API_BASE/v1/auth/complete" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"mode\": \"hybrid\",
    \"auth_session_id\": \"$HYBRID_SESSION_ID\",
    \"device_signature\": \"$HYBRID_SIGNATURE\"
  }" | jq .

# 12) Legacy direct SIM start (requires X-API-Key)
SIM_START_RESP=$(curl -s -X POST "$API_BASE/v1/sim/start" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $TIMBLE_API_KEY" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"msisdn\": \"917905968734\",
    \"action\": \"check\"
  }")
echo "$SIM_START_RESP" | jq .
SIM_SESSION_ID=$(echo "$SIM_START_RESP" | jq -r '.auth_session_id')

# 13) Legacy direct SIM complete
curl -s -X POST "$API_BASE/v1/sim/complete" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $TIMBLE_API_KEY" \
  -d "{
    \"auth_session_id\": \"$SIM_SESSION_ID\"
  }" | jq .

# 14) Device revoke
curl -s -X POST "$API_BASE/v1/device/revoke" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"device_binding_id\": \"$BINDING_ID\"
  }" | jq .

# 15) Final device check after revoke
curl -s -G "$API_BASE/v1/device/check" \
  --data-urlencode "client_id=$CLIENT_ID" \
  --data-urlencode "user_ref=$USER_REF" \
  --data-urlencode "device_id=$DEVICE_ID" | jq .

# ── New Device Verification Flow ──────────────────────────────
# When an unknown device calls /v1/auth/start, the server auto-triggers
# the approval flow and returns next_step=DEVICE_APPROVAL_REQUIRED.

NEW_DEVICE_ID="dev_new_999"
NEW_PUBLIC_KEY="LS0tLS1CRUd..."  # base64 ECDSA public key of new device

# 16) Unknown device attempts auth start (auto-triggers approval)
APPROVAL_RESP=$(curl -s -X POST "$API_BASE/v1/auth/start" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"action\": \"login\",
    \"mode\": \"device\",
    \"device_binding_id\": \"$BINDING_ID\",
    \"device_info\": {
      \"device_id\": \"$NEW_DEVICE_ID\",
      \"platform\": \"android\",
      \"device_model\": \"Pixel 8\",
      \"os_version\": \"14\"
    },
    \"public_key\": \"$NEW_PUBLIC_KEY\"
  }")
echo "$APPROVAL_RESP" | jq .
APPROVAL_ID=$(echo "$APPROVAL_RESP" | jq -r '.auth_session_id')
# Response: 202, next_step=DEVICE_APPROVAL_REQUIRED, auth_session_id=appr_...

# 17) Or call device-verify directly (alternative to auto-trigger)
curl -s -X POST "$API_BASE/v1/auth/device-verify" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"device_info\": {
      \"device_id\": \"$NEW_DEVICE_ID\",
      \"platform\": \"android\",
      \"device_model\": \"Pixel 8\",
      \"os_version\": \"14\"
    },
    \"public_key\": \"$NEW_PUBLIC_KEY\"
  }" | jq .

# 18) Main device polls for pending approvals
curl -s -G "$API_BASE/v1/auth/device-verify/pending" \
  --data-urlencode "client_id=$CLIENT_ID" \
  --data-urlencode "user_ref=$USER_REF" | jq .

# 19) Main device approves the new device
curl -s -X POST "$API_BASE/v1/auth/device-verify/respond" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"user_ref\": \"$USER_REF\",
    \"approval_id\": \"$APPROVAL_ID\",
    \"action\": \"approve\"
  }" | jq .

# 20) New device polls for approval status
curl -s -X POST "$API_BASE/v1/auth/device-verify/status" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$CLIENT_ID\",
    \"approval_id\": \"$APPROVAL_ID\"
  }" | jq .
# Response: status=APPROVED, message="Device approved — proceed with login"

# 21) New device can now call /v1/auth/start normally (device is now trusted)
```

## Production Differences

Before deploying to production:
1. Set `DEV_MODE=false`.
2. Delete `DEV_PRIVATE_KEY` from environment variables, it's not needed by the server in production. The app generates keys privately!
3. Ensure rate limiters on `.config` (nginx or ingress) are protecting `/v1/auth/start`.
4. Run HTTPS to encrypt all traffic containing Base64 payloads.
