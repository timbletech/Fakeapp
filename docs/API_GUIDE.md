# API Documentation: Timble Authentication & Deepfake Detection

Complete API reference for the Timble authentication system with integrated deepfake detection capabilities.

**Version:** 1.0.0  
**Base URL:** `http://localhost:8080` (development) | `https://api.example.com` (production)

## Table of Contents

- [Overview](#overview)
- [Authentication Modes](#authentication-modes)
- [Device Management](#device-management)
- [Authentication Endpoints](#authentication-endpoints)
- [Deepfake Detection](#deepfake-detection)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Quick Start Examples](#quick-start-examples)

---

## Overview

The Timble API provides unified authentication with three distinct modes:

| Mode | Description | Use Case |
|------|-------------|----------|
| **Device** | Cryptographic challenge-response using ECDSA | Single device verification |
| **SIM** | SIM-based verification via XConnect/Sekura provider | Telecom integration |
| **Hybrid** | SIM verification + device cryptographic signature | Maximum security |

Additionally, deepfake detection capabilities are provided for:
- Face recognition and liveness detection
- Voice authenticity verification
- Document tampering detection

---

## Device Management

Device management endpoints handle registration, verification, and lifecycle of device bindings.

### Check Device Status

**Endpoint:** `GET /v1/device/check`

Verify if a device is registered and retrieve its current status.

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `device_binding_id` | string | Yes | The unique device binding identifier |

**Response (200 OK):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "device_binding_id": "db_abc123def456",
  "status": "active",
  "client_id": "client_001",
  "user_ref": "user_xyz789",
  "created_at": "2024-01-15T10:30:00Z",
  "last_verified": "2024-04-01T14:22:15Z"
}
```

**cURL Example:**
```bash
curl -X GET "http://localhost:8080/v1/device/check?device_binding_id=db_abc123def456"
```

---

### Register Device

**Endpoint:** `POST /v1/device/register`

Register a new device binding for a user. If a binding already exists for this user, it will be replaced.

**Request Body:**
```json
{
  "client_id": "client_001",
  "user_ref": "user_example@bank.com",
  "public_key": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...\n-----END PUBLIC KEY-----",
  "device_info": {
    "device_id": "device_uuid_12345",
    "platform": "iOS",
    "app_version": "2.1.0",
    "device_model": "iPhone14,3",
    "os_version": "17.2",
    "ip_address": "192.168.1.100"
  }
}
```

**Response (200 OK):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-04-01T14:22:15Z",
  "client_id": "client_001",
  "device_binding_id": "db_abc123def456",
  "status": "active"
}
```

**Error Response (400):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "error": "invalid_key_format",
  "message": "Public key must be in PEM format"
}
```

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/v1/device/register" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client_001",
    "user_ref": "user@example.com",
    "public_key": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...\n-----END PUBLIC KEY-----",
    "device_info": {
      "device_id": "device_uuid_12345",
      "platform": "iOS",
      "app_version": "2.1.0"
    }
  }'
```

---

### Revoke Device

**Endpoint:** `POST /v1/device/revoke`

Revoke an active device binding, immediately invalidating any associated authentication sessions.

**Request Body:**
```json
{
  "device_binding_id": "db_abc123def456",
  "client_id": "client_001"
}
```

**Response (200 OK):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "device_binding_id": "db_abc123def456",
  "status": "revoked"
}
```

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/v1/device/revoke" \
  -H "Content-Type: application/json" \
  -d '{
    "device_binding_id": "db_abc123def456",
    "client_id": "client_001"
  }'
```

---

### Update Device

**Endpoint:** `PUT /v1/device/update`

Update device metadata and information (e.g., new app version, device model).

**Request Body:**
```json
{
  "device_binding_id": "db_abc123def456",
  "device_info": {
    "device_id": "device_uuid_12345",
    "platform": "iOS",
    "app_version": "2.2.0",
    "device_model": "iPhone14,3",
    "os_version": "17.3"
  }
}
```

**Response (200 OK):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "device_binding_id": "db_abc123def456",
  "status": "updated"
}
```

---

## Authentication Endpoints

### Start Authentication

**Endpoint:** `POST /v1/auth/start`

Initiate an authentication session in device, SIM, or hybrid mode. The mode is determined by:
1. The `mode` parameter in the request
2. Client policy configuration (`DEFAULT_AUTH_MODE`, `CLIENT_AUTH_MODES`)

**Request Body:**
```json
{
  "client_id": "client_001",
  "user_ref": "user@example.com",
  "action": "login",
  "mode": "hybrid",
  "device_binding_id": "db_abc123def456",
  "msisdn": "+1234567890",
  "device_info": {
    "device_id": "device_uuid_12345",
    "platform": "iOS",
    "app_version": "2.1.0"
  }
}
```

**Parameters:**
| Field | Type | Required | Mode(s) | Description |
|-------|------|----------|---------|-------------|
| `client_id` | string | Yes | All | Client identifier |
| `user_ref` | string | Yes | All | User reference/email |
| `action` | string | No | All | Action type (e.g., "login", "transaction") |
| `mode` | string | No | All | `device`, `sim`, or `hybrid` |
| `device_binding_id` | string | Yes | Device, Hybrid | Device binding ID |
| `msisdn` | string | Yes | SIM, Hybrid | Mobile number (+1234567890) |
| `device_info` | object | No | All | Device metadata |

**Response (200 OK) - Device Mode:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-04-01T14:22:15Z",
  "client_id": "client_001",
  "mode": "device",
  "auth_session_id": "auth_sess_xyz789",
  "next_step": "sign_challenge",
  "device": {
    "challenge_id": "chal_123abc",
    "challenge": "base64_encoded_32byte_random_challenge",
    "expires_in_seconds": 300
  },
  "status": "initiated"
}
```

**Response (200 OK) - SIM Mode:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-04-01T14:22:15Z",
  "client_id": "client_001",
  "mode": "sim",
  "auth_session_id": "auth_sess_xyz789",
  "next_step": "sim_approval",
  "sim": {
    "auth_session_id": "auth_sess_xyz789",
    "session_uri": "https://provider.example.com/confirm?session_id=...",
    "expires_in_seconds": 600,
    "instructions": "Open the SMS link and approve the authentication request"
  },
  "status": "initiated"
}
```

**Response (200 OK) - Hybrid Mode:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-04-01T14:22:15Z",
  "client_id": "client_001",
  "mode": "hybrid",
  "auth_session_id": "auth_sess_xyz789",
  "next_step": "dual_verification",
  "device": {
    "challenge_id": "chal_123abc",
    "challenge": "base64_encoded_challenge",
    "expires_in_seconds": 300
  },
  "sim": {
    "auth_session_id": "auth_sess_xyz789",
    "session_uri": "https://provider.example.com/confirm?session_id=...",
    "expires_in_seconds": 600
  },
  "status": "initiated"
}
```

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/v1/auth/start" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client_001",
    "user_ref": "user@example.com",
    "action": "login",
    "mode": "device",
    "device_binding_id": "db_abc123def456",
    "device_info": {
      "device_id": "device_uuid_12345",
      "platform": "iOS"
    }
  }'
```

---

### Complete Authentication

**Endpoint:** `POST /v1/auth/complete`

Complete an authentication session by providing:
- **Device mode:** ECDSA signature of the challenge
- **SIM mode:** Poll for SIM provider decision
- **Hybrid mode:** Both device signature AND SIM approval

**Request Body - Device Mode:**
```json
{
  "client_id": "client_001",
  "auth_session_id": "auth_sess_xyz789",
  "mode": "device",
  "challenge_id": "chal_123abc",
  "device_signature": "base64_ecdsa_signature_of_challenge"
}
```

**Request Body - SIM Mode:**
```json
{
  "client_id": "client_001",
  "auth_session_id": "auth_sess_xyz789",
  "mode": "sim"
}
```

**Request Body - Hybrid Mode:**
```json
{
  "client_id": "client_001",
  "auth_session_id": "auth_sess_xyz789",
  "mode": "hybrid",
  "challenge_id": "chal_123abc",
  "device_signature": "base64_ecdsa_signature_of_challenge"
}
```

**Response (200 OK) - ALLOW Decision:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-04-01T14:22:15Z",
  "client_id": "client_001",
  "mode": "device",
  "auth_session_id": "auth_sess_xyz789",
  "decision": "ALLOW",
  "reason_code": "success",
  "reason_message": "Authentication successful",
  "next_step": "complete",
  "attempts_remaining": 3,
  "auth_context_token": "ctx_token_abc123xyz789",
  "expires_in_seconds": 3600,
  "status": "completed"
}
```

**Response (200 OK) - DENY Decision:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-04-01T14:22:15Z",
  "client_id": "client_001",
  "mode": "device",
  "auth_session_id": "auth_sess_xyz789",
  "decision": "DENY",
  "reason_code": "invalid_signature",
  "reason_message": "Device signature verification failed",
  "next_step": "retry_or_fallback",
  "attempts_remaining": 2,
  "status": "failed"
}
```

**Response (202 Accepted) - Pending (SIM Still Deciding):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-04-01T14:22:15Z",
  "client_id": "client_001",
  "mode": "sim",
  "auth_session_id": "auth_sess_xyz789",
  "decision": "PENDING",
  "reason_message": "Waiting for SIM provider decision",
  "next_step": "retry_completion",
  "status": "pending"
}
```

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/v1/auth/complete" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client_001",
    "auth_session_id": "auth_sess_xyz789",
    "mode": "device",
    "challenge_id": "chal_123abc",
    "device_signature": "MEUCIQC46cKp3hKrVDJ1kBg9h/EqrKVJeHEtc7...(base64)"
  }'
```

---

### Verify Token

**Endpoint:** `POST /v1/auth/verify`

Validate an authentication context token received from a completed authentication session.

**Request Body:**
```json
{
  "client_id": "client_001",
  "auth_context_token": "ctx_token_abc123xyz789"
}
```

**Response (200 OK):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-04-01T14:22:15Z",
  "client_id": "client_001",
  "valid": true,
  "expires_in_seconds": 3500,
  "status": "valid"
}
```

**Response (200 OK) - Invalid Token:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-04-01T14:22:15Z",
  "client_id": "client_001",
  "valid": false,
  "status": "invalid"
}
```

---

## Hybrid Mode Routes (Aliases)

The following routes are convenience aliases for the standard auth endpoints:

- `POST /v1/hybrid/start` → Same as `POST /v1/auth/start` with `mode=hybrid`
- `POST /v1/hybrid/complete` → Same as `POST /v1/auth/complete` with `mode=hybrid`

---

## Deepfake Detection

### Face Detection

#### Analyze Image (Instant)

**Endpoint:** `POST /v1/face/image`

Detect face deepfakes in a single image with instant results.

**Request Format:** `multipart/form-data`

**Form Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | binary | Yes | Image file (JPEG, PNG, max 10MB) |

**Response (200 OK):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "verdict": "real",
  "confidence": 0.98,
  "analysis": {
    "face_detected": true,
    "quality_score": 0.95,
    "artifacts": []
  }
}
```

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/v1/face/image" \
  -F "file=@/path/to/image.jpg"
```

---

#### Submit Video (Async)

**Endpoint:** `POST /v1/face/video`

Submit a video file for asynchronous deepfake face detection analysis.

**Request Format:** `multipart/form-data`

**Form Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | binary | Yes | Video file (MP4, AVI, MOV, max 100MB) |

**Response (202 Accepted):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "status": "submitted",
  "execution_mode": "async",
  "task_id": "video_task_abc123xyz789",
  "poll_url": "/v1/face/video/video_task_abc123xyz789"
}
```

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/v1/face/video" \
  -F "file=@/path/to/video.mp4"
```

---

#### Poll Video Results

**Endpoint:** `GET /v1/face/video/{job_id}`

Retrieve the analysis results for a submitted video job.

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `job_id` | string | Yes | Task ID from video submission |

**Response (200 OK) - Complete:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "task_id": "video_task_abc123xyz789",
  "status": "completed",
  "verdict": "real",
  "confidence": 0.97,
  "frames_analyzed": 150,
  "duration_seconds": 5.0,
  "analysis": {
    "face_positions": [...],
    "liveness_confidence": 0.98,
    "artifacts_detected": false
  }
}
```

**Response (202 Accepted) - Still Processing:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "task_id": "video_task_abc123xyz789",
  "status": "processing",
  "message": "Estimated 30 seconds remaining"
}
```

**cURL Example:**
```bash
curl -X GET "http://localhost:8080/v1/face/video/video_task_abc123xyz789"
```

---

#### List Video Jobs

**Endpoint:** `GET /v1/face/video`

Retrieve a list of all submitted video analysis jobs.

**Response (200 OK):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "jobs": [
    {
      "task_id": "video_task_abc123",
      "status": "completed",
      "created_at": "2024-04-01T14:00:00Z",
      "updated_at": "2024-04-01T14:05:30Z"
    },
    {
      "task_id": "video_task_def456",
      "status": "processing",
      "created_at": "2024-04-01T14:20:00Z",
      "updated_at": "2024-04-01T14:22:15Z"
    }
  ]
}
```

---

### Voice Detection

#### Analyze Voice

**Endpoint:** `POST /v1/voice/analyze`

Analyze audio data for voice deepfake detection. Supports both synchronous and asynchronous modes.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `sync` | boolean | false | If true, wait for results synchronously |

**Request Body:**
```json
{
  "data": "base64_encoded_audio_data",
  "layers": ["spectral_analysis", "frequency_patterns"]
}
```

**Parameters:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `data` | string | Yes | Base64-encoded audio file |
| `layers` | array | No | Analysis layers to apply |

**Response (200 OK) - Async (default):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "status": "submitted",
  "execution_mode": "async",
  "task_id": "voice_task_abc123xyz789",
  "requested_layers": ["spectral_analysis"],
  "poll_url": "/v1/voice/analyze/voice_task_abc123xyz789"
}
```

**Response (200 OK) - Sync (with ?sync=true):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "verdict": "real",
  "confidence": 0.96,
  "duration_seconds": 8.5,
  "speakers_detected": 1,
  "analysis": {
    "pitch_stability": 0.92,
    "frequency_distribution": {...},
    "artifacts": []
  }
}
```

**cURL Example - Async:**
```bash
curl -X POST "http://localhost:8080/v1/voice/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "data": "SUQzBAAAAAAAI1NTUQAAAAAPAAADTGF2Zpk...(base64)"
  }'
```

**cURL Example - Sync:**
```bash
curl -X POST "http://localhost:8080/v1/voice/analyze?sync=true" \
  -H "Content-Type: application/json" \
  -d '{
    "data": "SUQzBAAAAAAAI1NTUQAAAAAPAAADTGF2Zpk...(base64)"
  }'
```

---

#### Poll Voice Results

**Endpoint:** `GET /v1/voice/analyze/{task_id}`

Retrieve the analysis results for a submitted voice task.

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_id` | string | Yes | Task ID from voice submission |

**Response (200 OK) - Complete:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "task_id": "voice_task_abc123xyz789",
  "status": "completed",
  "verdict": "real",
  "confidence": 0.96,
  "duration_seconds": 8.5,
  "speakers_detected": 1,
  "analysis": {
    "pitch_analysis": {...},
    "formant_frequencies": {...}
  }
}
```

**cURL Example:**
```bash
curl -X GET "http://localhost:8080/v1/voice/analyze/voice_task_abc123xyz789"
```

---

### Document Detection

#### Analyze Document

**Endpoint:** `POST /v1/deepfake/analyze`

Analyze document data for tampering detection. Supports both synchronous and asynchronous modes.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `sync` | boolean | false | If true, wait for results synchronously |

**Request Body:**
```json
{
  "data": "base64_encoded_document_data",
  "layers": ["metadata_analysis", "pixel_patterns"]
}
```

**Response (200 OK) - Async (default):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "status": "submitted",
  "execution_mode": "async",
  "task_id": "doc_task_abc123xyz789",
  "requested_layers": ["metadata_analysis"],
  "poll_url": "/v1/deepfake/analyze/doc_task_abc123xyz789"
}
```

**Response (200 OK) - Sync (with ?sync=true):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "verdict": "authentic",
  "confidence": 0.99,
  "tampering_regions": [],
  "analysis": {
    "metadata_intact": true,
    "compression_patterns": "original"
  }
}
```

**Response (200 OK) - Tampering Detected:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "verdict": "tampered",
  "confidence": 0.94,
  "tampering_regions": [
    {
      "region": "signature_area",
      "confidence": 0.95,
      "type": "replacement"
    }
  ],
  "analysis": {
    "altered_text_regions": 1,
    "metadata_modifications": true
  }
}
```

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/v1/deepfake/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "data": "JVBERi0xLjQK...(base64)"
  }'
```

---

#### Poll Document Results

**Endpoint:** `GET /v1/deepfake/analyze/{task_id}`

Retrieve the analysis results for a submitted document task.

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_id` | string | Yes | Task ID from document submission |

**Response (200 OK):**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "task_id": "doc_task_abc123xyz789",
  "status": "completed",
  "verdict": "authentic",
  "confidence": 0.99,
  "tampering_regions": [],
  "analysis": {...}
}
```

---

## Error Handling

All error responses use consistent format:

**Error Response Format:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "error": "error_code",
  "message": "Human-readable error description"
}
```

**Common Error Codes:**

| Code | HTTP | Description |
|------|------|-------------|
| `invalid_request` | 400 | Malformed request body or missing required fields |
| `invalid_key_format` | 400 | Public key not in valid PEM format |
| `invalid_signature` | 400 | Device signature verification failed |
| `invalid_token` | 400 | Invalid or malformed auth context token |
| `device_not_found` | 404 | Device binding does not exist |
| `session_expired` | 410 | Authentication session has expired |
| `session_not_found` | 404 | Authentication session not found |
| `task_not_found` | 404 | Analysis task not found |
| `internal_error` | 500 | Internal server error |
| `sim_provider_error` | 502 | SIM provider (XConnect/Sekura) error |

**Example Error Response:**
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "error": "invalid_signature",
  "message": "Device signature verification failed. Ensure the signature is a valid ECDSA P-256 signature of the challenge."
}
```

---

## Rate Limiting

API endpoints are rate-limited per client_id:

**Rate Limits:**
| Endpoint Category | Requests | Window |
|------------------|----------|--------|
| Authentication | 100 | 1 minute |
| Device Management | 50 | 1 minute |
| Deepfake Detection | 500 | 1 hour |

Rate limit information is returned in response headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1712074800
```

---

## Quick Start Examples

### Example 1: Device-Only Authentication Flow

1. **Register Device:**
```bash
curl -X POST "http://localhost:8080/v1/device/register" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "bankapp_ios",
    "user_ref": "john@bank.com",
    "public_key": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...\n-----END PUBLIC KEY-----",
    "device_info": {
      "device_id": "device_12345",
      "platform": "iOS",
      "app_version": "1.0.0"
    }
  }'
```

Response:
```json
{
  "device_binding_id": "db_abc123",
  "status": "active"
}
```

2. **Start Authentication:**
```bash
curl -X POST "http://localhost:8080/v1/auth/start" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "bankapp_ios",
    "user_ref": "john@bank.com",
    "mode": "device",
    "device_binding_id": "db_abc123",
    "device_info": {
      "device_id": "device_12345",
      "platform": "iOS"
    }
  }'
```

Response:
```json
{
  "auth_session_id": "auth_sess_xyz",
  "device": {
    "challenge_id": "chal_123",
    "challenge": "base64_challenge_data",
    "expires_in_seconds": 300
  }
}
```

3. **Complete Authentication:**
```bash
curl -X POST "http://localhost:8080/v1/auth/complete" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "bankapp_ios",
    "auth_session_id": "auth_sess_xyz",
    "mode": "device",
    "challenge_id": "chal_123",
    "device_signature": "base64_signature"
  }'
```

Response:
```json
{
  "decision": "ALLOW",
  "auth_context_token": "ctx_token_abc123",
  "expires_in_seconds": 3600
}
```

---

### Example 2: Face Liveness Detection

1. **Submit Image for Instant Analysis:**
```bash
curl -X POST "http://localhost:8080/v1/face/image" \
  -F "file=@selfie.jpg"
```

Response:
```json
{
  "verdict": "real",
  "confidence": 0.98,
  "analysis": {
    "face_detected": true,
    "quality_score": 0.95
  }
}
```

---

### Example 3: Document Tampering Detection

1. **Submit Document (Async):**
```bash
curl -X POST "http://localhost:8080/v1/deepfake/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "data": "base64_encoded_pdf"
  }'
```

Response:
```json
{
  "task_id": "doc_task_abc123",
  "status": "submitted"
}
```

2. **Poll Results:**
```bash
curl -X GET "http://localhost:8080/v1/deepfake/analyze/doc_task_abc123"
```

Response:
```json
{
  "verdict": "authentic",
  "confidence": 0.99,
  "tampering_regions": []
}
```

---

## Additional Resources

- **OpenAPI Specification:** [openapi-extended.yaml](openapi-extended.yaml)
- **Banking Integration Guide:** [bank_mobile_integration_guide.md](bank_mobile_integration_guide.md)
- **Configuration:** See `internal/config/config.go`
- **Crypto Details:** See `internal/crypto/crypto.go`

---

## Support

For technical support or integration questions:
- 📧 Email: support@example.com
- 🐛 Issues: Report via your support channel
- 📚 Documentation: https://docs.example.com

