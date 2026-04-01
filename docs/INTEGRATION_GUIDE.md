# Integration Guide: Timble Authentication & Deepfake Detection

Complete step-by-step guide for integrating Timble authentication and deepfake detection into your application.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Device-Only Flow](#device-only-flow)
4. [SIM-Only Flow](#sim-only-flow)
5. [Hybrid Flow](#hybrid-flow)
6. [Deepfake Detection Integration](#deepfake-detection-integration)
7. [Error Handling & Retry Logic](#error-handling--retry-logic)
8. [Testing & Debugging](#testing--debugging)

---

## Prerequisites

- **API Endpoint:** Base URL (http://localhost:8080 for local dev, http://3.108.166.176:8097 for production)
- **Client ID:** Your application's unique identifier (provided by Timble)
- **ECDSA Private Key:** For device mode (P-256 curve)
- **Encryption:** All endpoints support HTTPS (recommended for production)

---

## Initial Setup

### Step 1: Obtain Credentials

Contact Timble to receive:
- `client_id` - Your unique client identifier
- OAuth credentials (if using SIM mode)
- List of supported authentication modes

### Step 2: Generate ECDSA Keys (Device Mode)

Generate an ECDSA P-256 key pair for device signing:

**Using OpenSSL:**
```bash
# Generate private key
openssl ecparam -name prime256v1 -genkey -noout -out private.pem

# Export public key
openssl ec -in private.pem -pubout -out public.pem

# View public key in PEM format
cat public.pem
```

**Output (public.pem):**
```
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...
-----END PUBLIC KEY-----
```

### Step 3: Configure Your Client

Store securely:
- `client_id`
- `private_key` (device mode only - never expose)
- `api_base_url`

---

## Device-Only Flow

Best for: Single device, highest security

### Step 1: Register Device

When user installs app and logs in for first time:

```javascript
// Generate ECDSA key pair in secure storage on device
const publicKeyPEM = getPublicKeyFromDeviceSecureStorage();

const registerRequest = {
  client_id: "your_client_id",
  user_ref: "user@example.com",
  public_key: publicKeyPEM,
  device_info: {
    device_id: generateOrGetUUID(),
    platform: "iOS",
    app_version: "1.0.0",
    device_model: "iPhone14,3",
    os_version: "17.2"
  }
};

const response = await fetch("http://3.108.166.176:8097/v1/device/register", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(registerRequest)
});

const { device_binding_id, status } = await response.json();
// Store device_binding_id securely
```

**Success Response:**
```json
{
  "request_id": "req_123...",
  "device_binding_id": "db_abc123",
  "status": "active"
}
```

### Step 2: Start Authentication

When user attempts to authenticate:

```javascript
const startAuthRequest = {
  client_id: "your_client_id",
  user_ref: "user@example.com",
  action: "login",
  mode: "device",
  device_binding_id: "db_abc123", // from registration
  device_info: {
    device_id: getDeviceUUID(),
    platform: "iOS",
    app_version: "1.0.0"
  }
};

const response = await fetch("http://3.108.166.176:8097/v1/auth/start", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(startAuthRequest)
});

const {
  auth_session_id,
  device: { challenge_id, challenge, expires_in_seconds }
} = await response.json();

// Store for next step
sessionStorage.setItem("auth_session_id", auth_session_id);
sessionStorage.setItem("challenge_id", challenge_id);
```

**Response:**
```json
{
  "auth_session_id": "auth_sess_xyz",
  "device": {
    "challenge_id": "chal_123",
    "challenge": "base64_challenge_32_bytes",
    "expires_in_seconds": 300
  }
}
```

### Step 3: Sign Challenge

On the device, sign the challenge with the private key:

```javascript
// Decode base64 challenge
const challengeBytes = base64ToBytes(challenge);

// Sign with ECDSA private key (using device secure enclave if available)
const privateKey = getPrivateKeyFromSecureEnclave();
const signature = await cryptoLib.signChallenge(challengeBytes, privateKey);
const signatureBase64 = bytesToBase64(signature);
```

### Step 4: Complete Authentication

Send the signature to the server:

```javascript
const completeAuthRequest = {
  client_id: "your_client_id",
  auth_session_id: "auth_sess_xyz",
  mode: "device",
  challenge_id: "chal_123",
  device_signature: signatureBase64
};

const response = await fetch("http://3.108.166.176:8097/v1/auth/complete", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(completeAuthRequest)
});

const result = await response.json();

if (result.decision === "ALLOW") {
  // Success! User authenticated
  const { auth_context_token, expires_in_seconds } = result;
  localStorage.setItem("auth_token", auth_context_token);
  redirectToApp();
} else if (result.decision === "DENY") {
  showError("Authentication failed: " + result.reason_message);
} else {
  showError("Unexpected response status");
}
```

**Success Response:**
```json
{
  "decision": "ALLOW",
  "auth_context_token": "ctx_token_abc123",
  "expires_in_seconds": 3600,
  "status": "completed"
}
```

**Failure Response:**
```json
{
  "decision": "DENY",
  "reason_code": "invalid_signature",
  "reason_message": "Signature verification failed",
  "attempts_remaining": 2
}
```

### Step 5: Verify Token

Before allowing sensitive operations, verify the token is still valid:

```javascript
async function verifyAuthToken(token) {
  const response = await fetch("http://3.108.166.176:8097/v1/auth/verify", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      client_id: "your_client_id",
      auth_context_token: token
    })
  });

  const { valid, expires_in_seconds } = await response.json();

  if (valid && expires_in_seconds > 60) {
    return true; // Token is valid with >60s remaining
  } else if (valid) {
    // Token expiring soon, refresh authentication
    return await reauthenticate();
  } else {
    // Token invalid, require new authentication
    redirectToLogin();
    return false;
  }
}
```

---

## SIM-Only Flow

Best for: Telecom provider integration

### Step 1: Start SIM Authentication

```javascript
const startAuthRequest = {
  client_id: "your_client_id",
  user_ref: "user@example.com",
  action: "login",
  mode: "sim",
  msisdn: "+1234567890", // User's phone number
  device_info: {
    device_id: getDeviceUUID(),
    platform: "iOS"
  }
};

const response = await fetch("http://3.108.166.176:8097/v1/auth/start", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(startAuthRequest)
});

const {
  auth_session_id,
  sim: { session_uri, expires_in_seconds }
} = await response.json();

// Redirect user to SIM provider for approval
window.location.href = session_uri;
```

**Response:**
```json
{
  "auth_session_id": "auth_sess_xyz",
  "sim": {
    "session_uri": "https://provider.example.com/confirm?session_id=...",
    "expires_in_seconds": 600
  }
}
```

### Step 2: Poll for SIM Decision

After user is redirected back, poll for SIM provider's decision:

```javascript
async function pollSIMDecision(authSessionId) {
  const maxAttempts = 30;
  const pollInterval = 2000; // 2 seconds

  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    const response = await fetch("http://3.108.166.176:8097/v1/auth/complete", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        client_id: "your_client_id",
        auth_session_id: authSessionId,
        mode: "sim"
      })
    });

    if (response.status === 200) {
      // Decision made
      const result = await response.json();
      if (result.decision === "ALLOW") {
        // Authentication successful
        const { auth_context_token } = result;
        localStorage.setItem("auth_token", auth_context_token);
        return true;
      } else {
        // Decision was DENY
        showError("SIM authentication denied");
        return false;
      }
    } else if (response.status === 202) {
      // Still pending, wait and retry
      await new Promise(resolve => setTimeout(resolve, pollInterval));
    } else {
      // Error
      showError("Error polling SIM decision");
      return false;
    }
  }

  showError("SIM decision timeout");
  return false;
}
```

---

## Hybrid Flow

Best for: Maximum security (requires both device & SIM)

### Step 1: Start Hybrid Authentication

```javascript
const startAuthRequest = {
  client_id: "your_client_id",
  user_ref: "user@example.com",
  mode: "hybrid",
  device_binding_id: "db_abc123",
  msisdn: "+1234567890",
  device_info: { /* ... */ }
};

const response = await fetch("http://3.108.166.176:8097/v1/auth/start", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(startAuthRequest)
});

const {
  auth_session_id,
  device: { challenge_id, challenge },
  sim: { session_uri }
} = await response.json();
```

### Step 2: Parallel Actions Required

1. **Sign device challenge** (see device flow step 3)
2. **Redirect to SIM provider** (see SIM flow)

### Step 3: Complete Hybrid Authentication

Both device signature AND SIM approval are required:

```javascript
const completeAuthRequest = {
  client_id: "your_client_id",
  auth_session_id: "auth_sess_xyz",
  mode: "hybrid",
  challenge_id: "chal_123",
  device_signature: signatureBase64
};

const response = await fetch("http://3.108.166.176:8097/v1/auth/complete", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(completeAuthRequest)
});

const result = await response.json();

// Success only if BOTH conditions met:
// 1. Device signature is valid
// 2. SIM approval was granted
if (result.decision === "ALLOW") {
  // Both authentications successful
}
```

---

## Deepfake Detection Integration

### Face Liveness Detection

#### Instant Image Analysis

```javascript
async function checkFaceLiveness(imageFile) {
  const formData = new FormData();
  formData.append('file', imageFile);

  const response = await fetch("http://3.108.166.176:8097/v1/face/image", {
    method: "POST",
    body: formData
  });

  const result = await response.json();
  // { verdict: "real"|"fake"|"unclear", confidence: 0.98 }
  return result.verdict === "real" && result.confidence > 0.9;
}
```

#### Video Analysis

```javascript
async function analyzeVideoForDeepfake(videoFile) {
  const formData = new FormData();
  formData.append('file', videoFile);

  // Step 1: Submit for async processing
  const submitResponse = await fetch("http://3.108.166.176:8097/v1/face/video", {
    method: "POST",
    body: formData
  });

  const { task_id } = await submitResponse.json();

  // Step 2: Poll for results
  let complete = false;
  let result = null;

  while (!complete) {
    const pollResponse = await fetch(
      `http://3.108.166.176:8097/v1/face/video/${task_id}`
    );

    if (pollResponse.status === 200) {
      result = await pollResponse.json();
      complete = true;
    } else if (pollResponse.status === 202) {
      // Still processing
      await new Promise(resolve => setTimeout(resolve, 1000));
    } else {
      throw new Error("Analysis failed");
    }
  }

  return result;
}
```

### Voice Biometric Detection

```javascript
async function analyzeVoiceSync(audioBase64) {
  // Synchronous analysis
  const response = await fetch(
    "http://3.108.166.176:8097/v1/voice/analyze?sync=true",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        data: audioBase64,
        layers: ["spectral_analysis", "frequency_patterns"]
      })
    }
  );

  const result = await response.json();
  // { verdict: "real"|"fake"|"unclear", confidence: 0.96 }
  return result;
}

async function analyzeVoiceAsync(audioBase64) {
  // Asynchronous analysis
  const submitResponse = await fetch(
    "http://3.108.166.176:8097/v1/voice/analyze",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ data: audioBase64 })
    }
  );

  const { task_id } = await submitResponse.json();

  // Poll for results
  const pollResponse = await fetch(
    `http://3.108.166.176:8097/v1/voice/analyze/${task_id}`
  );

  return await pollResponse.json();
}
```

### Document Tampering Detection

```javascript
async function detectDocumentTampering(documentBase64) {
  const response = await fetch(
    "http://3.108.166.176:8097/v1/deepfake/analyze?sync=true",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        data: documentBase64,
        layers: ["metadata_analysis", "pixel_patterns", "font_analysis"]
      })
    }
  );

  const result = await response.json();
  // { verdict: "authentic"|"tampered"|"unclear", confidence: 0.99 }

  if (result.verdict === "tampered") {
    console.log("Tampering detected in regions:", result.tampering_regions);
  }

  return result;
}
```

---

## Error Handling & Retry Logic

### Generic Error Handler

```javascript
async function handleAPIError(error, operation) {
  if (error.code === "invalid_signature") {
    // Device signature invalid - ask user to try again
    showError("Authentication failed. Please try again.");
    return "retry";
  } else if (error.code === "device_not_found") {
    // Device not registered - need re-registration
    redirectToDeviceRegistration();
    return "require_registration";
  } else if (error.code === "session_expired") {
    // Session timed out
    redirectToLogin();
    return "require_reauth";
  } else if (error.code === "sim_provider_error") {
    // External SIM provider error - retry with backoff
    return "retry_with_backoff";
  } else if (error.code === "internal_error") {
    // Server error - retry with exponential backoff
    return "retry_with_backoff";
  }
  return "unknown_error";
}
```

### Retry with Exponential Backoff

```javascript
async function retryWithBackoff(fn, maxAttempts = 3) {
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (error) {
      if (attempt === maxAttempts) throw error;

      // Exponential backoff: 2^n seconds
      const waitTime = Math.pow(2, attempt) * 1000;
      console.log(`Retry attempt ${attempt}, waiting ${waitTime}ms`);
      await new Promise(resolve => setTimeout(resolve, waitTime));
    }
  }
}

// Usage
const result = await retryWithBackoff(async () => {
  return await completeAuthentication(authSessionId);
});
```

---

## Testing & Debugging

### Test Device Registration

```bash
# Register test device
curl -X POST http://localhost:8080/v1/device/register \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "test_client",
    "user_ref": "test@example.com",
    "public_key": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE[...]\n-----END PUBLIC KEY-----",
    "device_info": {
      "device_id": "test_device_001",
      "platform": "iOS",
      "app_version": "1.0.0"
    }
  }' | jq .
```

### Test Authentication Flow

```bash
# 1. Start auth
AUTH_START=$(curl -X POST http://localhost:8080/v1/auth/start \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "test_client",
    "user_ref": "test@example.com",
    "mode": "device",
    "device_binding_id": "db_abc123",
    "device_info": {
      "device_id": "test_device_001",
      "platform": "iOS"
    }
  }')

AUTH_SESSION_ID=$(echo $AUTH_START | jq -r '.auth_session_id')
CHALLENGE=$(echo $AUTH_START | jq -r '.device.challenge')
CHALLENGE_ID=$(echo $AUTH_START | jq -r '.device.challenge_id')

echo "Auth Session: $AUTH_SESSION_ID"
echo "Challenge: $CHALLENGE"
echo "Challenge ID: $CHALLENGE_ID"

# 2. Sign challenge (in your app or test signing tool)
SIGNATURE=$(./sign_challenge $CHALLENGE)

# 3. Complete auth
curl -X POST http://localhost:8080/v1/auth/complete \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"test_client\",
    \"auth_session_id\": \"$AUTH_SESSION_ID\",
    \"mode\": \"device\",
    \"challenge_id\": \"$CHALLENGE_ID\",
    \"device_signature\": \"$SIGNATURE\"
  }" | jq .
```

### Debug Token Issues

```bash
# Verify token
curl -X POST http://localhost:8080/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "test_client",
    "auth_context_token": "ctx_token_abc123"
  }' | jq .
```

### Test Deepfake Detection

```bash
# Test face image detection
curl -X POST http://localhost:8080/v1/face/image \
  -F "file=@test_image.jpg" | jq .

# Test document analysis
curl -X POST "http://localhost:8080/v1/deepfake/analyze?sync=true" \
  -H "Content-Type: application/json" \
  -d '{
    "data": "base64_encoded_document"
  }' | jq .
```

### Debugging Data Encoding

```javascript
// Convert image to base64
const imageToBase64 = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => {
      const base64 = reader.result.split(',')[1];
      resolve(base64);
    };
    reader.onerror = reject;
  });
};

// Convert audio to base64
const audioToBase64 = (blob) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(blob);
    reader.onload = () => resolve(reader.result.split(',')[1]);
    reader.onerror = reject;
  });
};
```

---

## Best Practices

1. **Never expose private keys** - Keep ECDSA private keys in secure device storage
2. **Implement timeout handling** - SIM auth can take time; use polling with timeouts
3. **Add retry logic** - Network issues happen; implement exponential backoff
4. **Cache device binding ID** - Store securely after registration to avoid re-registering
5. **Validate tokens before use** - Call `/v1/auth/verify` before sensitive operations
6. **Log important events** - Track authentication successes, failures, and timestamps
7. **Monitor rate limits** - Check `X-RateLimit-Remaining` headers
8. **Use HTTPS in production** - Always use encrypted connections for sensitive data

---

## Support

For integration help, contact: integration@example.com

