# API Quick Reference

## Base URL
```
Development:  http://localhost:8080
Production:   https://api.example.com
```

## Device Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/device/check` | Check device binding status |
| POST | `/v1/device/register` | Register a new device |
| POST | `/v1/device/revoke` | Revoke a device binding |
| PUT | `/v1/device/update` | Update device information |

## Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/auth/start` | Start auth session (device/sim/hybrid) |
| POST | `/v1/auth/complete` | Complete auth session |
| POST | `/v1/auth/verify` | Verify auth context token |
| POST | `/v1/hybrid/start` | Alias for hybrid start |
| POST | `/v1/hybrid/complete` | Alias for hybrid complete |
| POST | `/v1/sim/start` | Direct SIM API (internal) |
| POST | `/v1/sim/complete` | Direct SIM completion (internal) |

## Deepfake Detection - Face

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/face/image` | Detect deepfake in image (instant) |
| POST | `/v1/face/video` | Submit video for analysis (async) |
| GET | `/v1/face/video` | List submitted video jobs |
| GET | `/v1/face/video/{job_id}` | Poll video analysis results |

## Deepfake Detection - Voice

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/voice/analyze` | Analyze audio (sync or async) |
| GET | `/v1/voice/analyze/{task_id}` | Poll voice analysis results |

## Deepfake Detection - Document

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/deepfake/analyze` | Analyze document (sync or async) |
| GET | `/v1/deepfake/analyze/{task_id}` | Poll document analysis results |

---

## Common Request Examples

### Start Device Authentication
```bash
curl -X POST http://localhost:8080/v1/auth/start \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client_001",
    "user_ref": "user@example.com",
    "mode": "device",
    "device_binding_id": "db_abc123"
  }'
```

### Complete Device Authentication
```bash
curl -X POST http://localhost:8080/v1/auth/complete \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client_001",
    "auth_session_id": "auth_sess_xyz789",
    "challenge_id": "chal_123",
    "device_signature": "base64_signature"
  }'
```

### Face Image Detection
```bash
curl -X POST http://localhost:8080/v1/face/image \
  -F "file=@image.jpg"
```

### Voice Analysis (Async)
```bash
curl -X POST http://localhost:8080/v1/voice/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "data": "base64_audio_data"
  }'
```

### Voice Analysis (Sync)
```bash
curl -X POST "http://localhost:8080/v1/voice/analyze?sync=true" \
  -H "Content-Type: application/json" \
  -d '{
    "data": "base64_audio_data"
  }'
```

### Poll Results
```bash
curl -X GET http://localhost:8080/v1/voice/analyze/voice_task_abc123
```

---

## Response Status Codes

| Code | Meaning |
|------|---------|
| 200 | OK - Request successful |
| 202 | Accepted - Processing async request |
| 400 | Bad Request - Validation error |
| 404 | Not Found - Resource doesn't exist |
| 410 | Gone - Session/token expired |
| 500 | Internal Server Error |
| 502 | Bad Gateway - Upstream provider error |

---

## Common Error Codes

| Code | HTTP | Description |
|------|------|-------------|
| `invalid_request` | 400 | Malformed request |
| `invalid_key_format` | 400 | Invalid key format |
| `invalid_signature` | 400 | Signature verification failed |
| `device_not_found` | 404 | Device doesn't exist |
| `session_expired` | 410 | Session has expired |
| `task_not_found` | 404 | Analysis task not found |
| `sim_provider_error` | 502 | SIM provider error |

---

## Response Example

**Success Response:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "status": "succeeded",
  "data": {...}
}
```

**Error Response:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "error": "error_code",
  "message": "Error description"
}
```

---

## Authentication Modes

### Device Mode
- Requires: `device_binding_id`
- Uses: ECDSA P-256 challenge-response
- Best for: Single device verification

### SIM Mode
- Requires: `msisdn` (phone number)
- Uses: XConnect/Sekura provider
- Best for: Telecom integration

### Hybrid Mode
- Requires: `device_binding_id` + `msisdn`
- Uses: SIM approval + device signature
- Best for: Maximum security

---

## Deepfake Analysis Verdicts

| Verdict | Meaning |
|---------|---------|
| `real` | Content appears authentic (face/voice) |
| `fake` | Deepfake detected |
| `unclear` | Unable to determine (low confidence) |
| `authentic` | Document appears unaltered |
| `tampered` | Document tampering detected |

---

## Rate Limits

- **Authentication:** 100 requests/minute per client
- **Device Management:** 50 requests/minute per client
- **Deepfake Detection:** 500 requests/hour per client

Rate limit headers in response:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1712074800
```

---

## Useful Links

- **Full API Guide:** [API_GUIDE.md](API_GUIDE.md)
- **OpenAPI Spec:** [openapi-extended.yaml](openapi-extended.yaml)
- **Banking Integration:** [bank_mobile_integration_guide.md](bank_mobile_integration_guide.md)

---

## SDK Resources

When available, SDKs generated from OpenAPI spec:
- **JavaScript/TypeScript:** Install from npm
- **Python:** Install from pip
- **Java:** Maven dependency
- **Go:** Go module

Generate client libraries:
```bash
# Using openapi-generator
openapi-generator generate -i openapi-extended.yaml -g go -o ./generated-client
```

