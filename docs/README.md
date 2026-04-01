# Timble API Documentation

Complete API documentation and integration guides for the Timble authentication system with deepfake detection capabilities.

## 📚 Documentation Structure

### 1. **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** ⭐ Start Here
Quick lookup for all API endpoints, common curl examples, and status codes. Perfect for developers who need a crash course.

**Includes:**
- All endpoint URLs at a glance
- Common curl examples
- HTTP status codes reference
- Response examples
- Rate limits summary

### 2. **[API_GUIDE.md](API_GUIDE.md)** 📖 Complete Reference
Comprehensive guide covering all API endpoints with detailed request/response examples, parameters, and error codes.

**Covers:**
- Device Management (check, register, revoke, update)
- Authentication Flows (device, SIM, hybrid modes)
- Token Verification
- Deepfake Detection (face, voice, document)
- Error handling & codes
- Rate limiting details

### 3. **[INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md)** 🚀 How-To Guide
Step-by-step integration guide with code examples in JavaScript/TypeScript showing how to implement each flow.

**Includes:**
- Prerequisites & setup
- Device-only authentication (complete flow)
- SIM-only authentication (complete flow)
- Hybrid authentication (complete flow)
- Deepfake detection integration examples
- Error handling & retry logic
- Testing & debugging commands
- Best practices

### 4. **[openapi-extended.yaml](openapi-extended.yaml)** 🔧 Technical Spec
Full OpenAPI 3.0.3 specification covering all endpoints, request/response schemas, and security definitions.

**Features:**
- All paths and operations defined
- Complete schema definitions
- Request/response examples
- Security schemes (API key auth)
- Tags for endpoint organization
- Generate client SDKs from this file

### 5. **[bank_mobile_integration_guide.md](bank_mobile_integration_guide.md)** 🏦 Domain-Specific
Banking and mobile app specific integration guide (if applicable to your use case).

---

## 🎯 Quick Navigation

### I want to...

| Goal | Start with |
|------|-----------|
| Get started quickly | [QUICK_REFERENCE.md](QUICK_REFERENCE.md) |
| Understand all endpoints | [API_GUIDE.md](API_GUIDE.md) |
| Implement authentication | [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) |
| Generate SDK client | [openapi-extended.yaml](openapi-extended.yaml) |
| Integrate with banking | [bank_mobile_integration_guide.md](bank_mobile_integration_guide.md) |
| Look up error codes | [API_GUIDE.md](API_GUIDE.md#error-handling) |
| Check rate limits | [QUICK_REFERENCE.md](QUICK_REFERENCE.md#rate-limits) |

---

## 🌟 Key Features

### Authentication Modes

| Mode | Best For | Complexity |
|------|----------|-----------|
| **Device** | Single device verification | Low |
| **SIM** | Telecom provider integration | Medium |
| **Hybrid** | Maximum security (both required) | High |

### Deepfake Detection

| Type | Use Case |
|------|----------|
| **Face** | Liveness detection, identity verification |
| **Voice** | Voice biometric, speaker verification |
| **Document** | Document tampering detection, KYC |

---

## 📋 API Endpoints Summary

### Device Management
```
GET  /v1/device/check          - Check device status
POST /v1/device/register       - Register device
POST /v1/device/revoke         - Revoke device
PUT  /v1/device/update         - Update device info
```

### Authentication
```
POST /v1/auth/start            - Start authentication
POST /v1/auth/complete         - Complete authentication
POST /v1/auth/verify           - Verify token
POST /v1/hybrid/start          - Start hybrid auth
POST /v1/hybrid/complete       - Complete hybrid auth
```

### Deepfake Detection
```
POST /v1/face/image            - Detect face in image
POST /v1/face/video            - Analyze video (async)
GET  /v1/face/video/{job_id}   - Poll video results
POST /v1/voice/analyze         - Analyze voice
GET  /v1/voice/analyze/{id}    - Poll voice results
POST /v1/deepfake/analyze      - Analyze document
GET  /v1/deepfake/analyze/{id} - Poll document results
```

---

## 🚀 Getting Started

### Prerequisites
- API endpoint URL (development or production)
- `client_id` (provided by Timble)
- ECDSA P-256 key pair (for device mode)

### First Steps

1. **Read** [QUICK_REFERENCE.md](QUICK_REFERENCE.md) (5 min)
2. **Review** your auth mode requirements
3. **Follow** [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) for your use case
4. **Reference** [API_GUIDE.md](API_GUIDE.md) for detailed parameter info
5. **Test** with curl examples
6. **Deploy** to production with HTTPS

### Generate API Client

```bash
# Using openapi-generator
openapi-generator generate \
  -i openapi-extended.yaml \
  -g go \
  -o ./timble-client-go

# For other languages, change -g option:
# -g typescript-fetch
# -g python
# -g java
```

---

## 🔐 Security Considerations

### Best Practices

- ✅ Keep ECDSA private keys in secure device storage (never expose)
- ✅ Always use HTTPS for API calls in production
- ✅ Validate auth tokens before sensitive operations (`/v1/auth/verify`)
- ✅ Implement timeout handling for SIM auth flows
- ✅ Use exponential backoff for retry logic
- ✅ Monitor rate limits and plan accordingly
- ✅ Log authentication events for audit trails

### Never Do

- ❌ Store private keys in code or config files
- ❌ Send unencrypted requests over HTTP to production
- ❌ Display full error messages to end users
- ❌ Retry indefinitely without backoff
- ❌ Cache auth tokens beyond their expiration

---

## 📊 Response Codes

### Success Codes
- `200 OK` - Request successful
- `202 Accepted` - Async processing accepted (polling required)

### Client Error Codes
- `400 Bad Request` - Invalid request format or parameters
- `404 Not Found` - Resource doesn't exist
- `410 Gone` - Session/token expired

### Server Error Codes
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Upstream provider error (SIM)

---

## 💡 Common Integration Scenarios

### Scenario 1: Device-Only Login
1. Register device once during app setup
2. On login, call `/v1/auth/start` with device mode
3. Sign challenge with device private key
4. Call `/v1/auth/complete` with signature
5. Receive `auth_context_token`

→ See [INTEGRATION_GUIDE.md - Device-Only Flow](INTEGRATION_GUIDE.md#device-only-flow)

### Scenario 2: SIM-Based Authentication
1. Call `/v1/auth/start` with SIM mode
2. Redirect user to SIM provider URI
3. After user approves, poll `/v1/auth/complete`
4. Receive `auth_context_token`

→ See [INTEGRATION_GUIDE.md - SIM-Only Flow](INTEGRATION_GUIDE.md#sim-only-flow)

### Scenario 3: High-Security Hybrid
1. Call `/v1/auth/start` with hybrid mode
2. Parallel: sign device challenge + redirect to SIM provider
3. After both complete, call `/v1/auth/complete` with signature
4. Server verifies both device signature AND SIM approval

→ See [INTEGRATION_GUIDE.md - Hybrid Flow](INTEGRATION_GUIDE.md#hybrid-flow)

### Scenario 4: Face Verification in KYC
1. Capture selfie on device
2. Call `/v1/face/image` for instant liveness check
3. If real, proceed with KYC; if fake, reject

→ See [INTEGRATION_GUIDE.md - Face Liveness](INTEGRATION_GUIDE.md#face-liveness-detection)

### Scenario 5: Document Upload Verification
1. User uploads document (ID, passport, etc.)
2. Call `/v1/deepfake/analyze?sync=true`
3. Check verdict field for tampering
4. Accept if "authentic", reject if "tampered"

→ See [INTEGRATION_GUIDE.md - Document Detection](INTEGRATION_GUIDE.md#document-tampering-detection)

---

## 🧪 Testing

### Test Environment

```
Base URL: http://localhost:8080
Test Client ID: test_client
```

### Quick Test

```bash
# Test API is running
curl http://localhost:8080/v1/device/check?device_binding_id=test

# Full test sequence - see INTEGRATION_GUIDE.md Testing & Debugging section
```

---

## 📞 Support

### Documentation
- **Full Guide:** [API_GUIDE.md](API_GUIDE.md)
- **Integration Help:** [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md)
- **Quick Lookup:** [QUICK_REFERENCE.md](QUICK_REFERENCE.md)

### Contact
- 📧 **Email:** support@example.com
- 🐛 **Report Issues:** [Submit via support portal]
- 📚 **API Status:** https://status.example.com

---

## 📈 Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2024-04-01 | Initial release with device, SIM, hybrid auth + deepfake detection |

---

## 📄 License

API documentation and examples provided as-is. For licensed usage terms, see your service agreement.

---

## 🔗 Related Files

- [Original OpenAPI spec](openapi.yaml) - Legacy/original specification
- [API Flow Diagram](api-flow.md) - Visual flow diagrams
- [README](../README.md) - Project overview
- [Deployment Guide](../deployment.md) - Deployment instructions

---

**Last Updated:** April 1, 2024

For the latest documentation, always refer to the official docs directory.
