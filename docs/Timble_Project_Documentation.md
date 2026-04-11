---
title: "Timble - Multi-Factor Authentication & Deepfake Detection Platform"
subtitle: "Complete Project Documentation"
date: "April 2026"
geometry: margin=2.5cm
fontsize: 11pt
toc: true
toc-depth: 3
header-includes:
  - \usepackage{booktabs}
  - \usepackage{longtable}
  - \usepackage{array}
  - \usepackage{fancyhdr}
  - \pagestyle{fancy}
  - \fancyhead[L]{Timble Platform}
  - \fancyhead[R]{Project Documentation}
  - \fancyfoot[C]{\thepage}
---

\newpage

# Product Identity

**Product Name:** Timble

**Tagline:** Multi-Factor Authentication and AI-Powered Deepfake Detection Platform

**What Timble Is:** Timble is a unified API gateway that combines cryptographic device authentication, telecom-level SIM fraud detection, and real-time AI-powered deepfake detection into a single platform purpose-built for banking and financial institutions. It verifies not just *what you have* (device, SIM) but also *that your SIM hasn't been hijacked* (SIM swap check), *that the device is genuine* (device ID verification), and *who you are* (face liveness, voice authenticity, document integrity).

**Core Capabilities:**

| Capability | Description |
|:-----------|:------------|
| Device Authentication | ECDSA P-256 challenge-response using device secure enclave |
| SIM Swap Detection | Telecom-level verification to detect recently swapped SIM cards |
| Device ID Verification | Matches device identifier against registered binding to prevent cloning |
| Face Image Detection | Tri-detector ensemble + 26 heuristic signals + CLIP + deepfake classifier |
| Face Video Detection | Temporal frame-by-frame deepfake analysis with vote aggregation |
| Voice Deepfake Detection | VARE (Voice Authenticity Risk Engine) for synthetic speech detection |
| Document Tampering | AI-powered forgery and manipulation detection for identity documents |
| Liveness Detection | Blink, mouth, and head-nod challenges to prevent replay attacks |
| Hybrid Authentication | Parallel device + SIM verification in a single authentication flow |

**Deployment:** Live production server at `3.108.166.176:8097` with Nginx TLS termination, PostgreSQL database, and distributed AI microservices.

\newpage

# Problem Statement

## The Threat Landscape

Financial institutions face a rapidly escalating convergence of identity fraud vectors that existing single-purpose solutions fail to address:

1. **Deepfake-Driven Identity Fraud** --- Generative AI tools (Stable Diffusion, Midjourney, DALL-E, voice cloning services) have made it trivially easy to create photorealistic fake faces, synthetic voices, and forged documents. Traditional KYC processes that rely on photo matching, voice verification, or document inspection are increasingly defeated by AI-generated content that passes human review.

2. **SIM Swap Account Takeover** --- Attackers port a victim's phone number to a new SIM card, intercepting OTP codes and SMS-based authentication. SIM swap fraud is one of the fastest-growing attack vectors in Indian banking, with losses running into crores annually. OTP-based authentication alone provides no protection against ported numbers.

3. **Device Cloning and Spoofing** --- Sophisticated attackers clone device identifiers or use emulators to impersonate trusted devices, bypassing device-fingerprint-based security. Without cryptographic device binding, banks cannot distinguish a genuine device from a spoofed one.

4. **Presentation Attacks** --- Photos, videos, or 3D masks held up to cameras during video KYC or selfie verification defeat basic face matching systems. Without liveness detection, any static or replayed media can pass as a live user.

5. **Fragmented Security Stack** --- Banks currently integrate 4--6 separate vendors for authentication (OTP provider), device binding (SDK vendor), face matching (KYC vendor), liveness (separate SDK), document verification (OCR vendor), and fraud detection (rule engine). Each integration adds latency, cost, failure points, and compliance complexity.

## The Gap

No single platform today combines cryptographic device authentication, telecom SIM fraud detection, and multi-modal AI deepfake detection (face, voice, document) into a unified API. Banks are forced to stitch together disparate solutions, resulting in:

- **Integration overhead** --- Multiple SDKs, APIs, and vendor contracts
- **Latency accumulation** --- Sequential calls to 4--6 services per authentication
- **Blind spots** --- SIM swap goes unchecked during face verification; document forgery is not validated during device auth
- **Compliance burden** --- Audit trails scattered across multiple vendors

## Timble's Answer

Timble eliminates this fragmentation by providing a single API endpoint that orchestrates device cryptography, SIM verification, and AI deepfake detection in parallel --- delivering a comprehensive identity assurance verdict in a single call, with a unified audit trail for regulatory compliance.

\newpage

# What It Demonstrates

## 1. Unified Multi-Factor Identity Assurance

Timble demonstrates that cryptographic authentication, telecom verification, and AI-powered fraud detection can be combined into a single cohesive platform rather than requiring separate vendor integrations. A single `POST /v1/auth/start` call with `mode: "hybrid"` initiates device challenge, SIM swap check, and device ID verification simultaneously.

## 2. Production-Grade AI Deepfake Detection

The face detection pipeline demonstrates a multi-layered AI approach to deepfake detection:

- **Tri-detector ensemble** (Dlib + RetinaFace + MediaPipe) ensures faces are detected regardless of angle, scale, or occlusion.
- **26+ heuristic artifact signals** catch statistical anomalies that AI-generated images exhibit (noise uniformity, frequency artifacts, symmetry, 3D geometry).
- **CLIP zero-shot classification** with 5-prompt ensembling provides robust AI vs Real scoring without task-specific training data.
- **Dedicated deepfake classifier** with temperature-scaled calibration prevents overconfident false positives.
- **Weighted consensus voting** requires agreement across multiple independent signals before flagging, dramatically reducing false positive rates compared to single-model approaches.

## 3. Real-Time SIM Fraud Prevention

Live integration with telecom SIM verification APIs demonstrates real-time detection of SIM swap attacks --- a capability that OTP-based systems fundamentally cannot provide. The system detects whether a phone number's SIM card has been recently ported before granting authentication.

## 4. Temporal Video Deepfake Analysis

The video analysis pipeline demonstrates frame-by-frame temporal deepfake detection using an Xception + LSTM architecture that captures both spatial artifacts (per-frame) and temporal inconsistencies (across frames) --- detecting deepfake videos that may pass single-frame analysis.

## 5. Multi-Modal Fraud Coverage

Timble demonstrates detection across all three identity fraud modalities:

- **Visual** --- Face image and video deepfake detection with liveness challenges
- **Audio** --- Voice cloning and synthetic speech detection via VARE
- **Document** --- Tampering, manipulation, and forgery detection for identity documents

## 6. Graceful Degradation Under Partial Failure

The weighted consensus engine demonstrates production resilience: if any individual ML model (CLIP, deepfake classifier, or heuristics) becomes unavailable, the system automatically redistributes weights among available models and continues operating with reduced but functional detection capability.

## 7. Zero-Framework High-Performance Architecture

The Go backend demonstrates that a production banking API can be built using only the standard library (`net/http`), achieving minimal memory footprint, fast startup, and zero external dependency vulnerabilities --- critical for banking security compliance.

\newpage

# Innovation Involved

Multi-factor authentication combining **Device-based ECDSA cryptographic challenge-response**, **SIM Swap detection**, **Device ID verification**, and **AI-powered deepfake detection** (face, voice, and document) into a single unified API platform for banking and financial institutions. The innovation lies in merging traditional authentication with **telecom-level SIM fraud detection**, **device binding verification**, and **real-time AI fraud detection** --- enabling banks to verify not just *what you have* (device, SIM) but also *that your SIM hasn't been hijacked* (SIM swap check), *that the device is genuine* (device ID verification), and *who you are* (face liveness, voice authenticity, document integrity) through a single API gateway.

The face detection system employs a **tri-detector ensemble** (Dlib HOG + RetinaFace + MediaPipe) combined with **26+ heuristic artifact signals**, **CLIP multi-prompt zero-shot classification**, and a **dedicated deepfake binary classifier** --- all fused through a weighted consensus voting engine that prevents single-model false positives.

\newpage

# Technology Used

| Layer | Technology |
|:------|:-----------|
| Authentication Backend | Go 1.22, stdlib `net/http` (zero-framework), Port 8097 |
| Face Detection Backend | Python, FastAPI 0.110.0+, Uvicorn, Port 8001 |
| Database | PostgreSQL (localhost:5432) |
| Cryptography | ECDSA P-256 Elliptic Curve (SHA-256 signing, 32-byte random challenges) |
| Telecom Integration | SIM-based phone number verification over mobile network (SIM Swap Check and Device ID Verification) |
| Face Detection Models | Dlib 19.24 (HOG + 68-point landmarks), RetinaFace (multi-scale pyramid), MediaPipe 0.10 (468 3D landmarks + blendshapes) |
| AI Classification Models | OpenAI CLIP ViT-Base-Patch32 (zero-shot, 45% weight), dima806/deepfake\_vs\_real\_image\_detection (binary classifier, 35% weight) |
| ML Framework | PyTorch 2.1+, Hugging Face Transformers 4.38+ |
| Image Processing | OpenCV 4.8+ (Laplacian, FFT, Canny), Pillow 10.0+ (EXIF, ELA), SciPy 1.11+ (FFT, spatial analysis), NumPy 1.24+ |
| Deepfake Detection Service | Remote AI server at `http://43.204.41.9` |
| Voice Detection Service | Local microservice at `localhost:8096` |
| Reverse Proxy / TLS | Nginx |
| Production Server | Linux (`3.108.166.176:8097`) |
| Go Dependencies | `lib/pq`, `google/uuid`, `godotenv` |
| Version Control | Git |

\newpage

# Development Undertaken in the Solution During Development Phase

## 1. Three-Mode Authentication Engine
Designed and implemented Device-only, SIM-only, and Hybrid authentication modes with a unified orchestration layer that coordinates both factors in parallel during hybrid auth.

## 2. ECDSA P-256 Cryptographic System
Built a complete challenge-response system using industry-standard 256-bit elliptic curve cryptography with 32-byte random challenge generation, ASN.1 signature encoding, and secure key storage.

## 3. SIM Swap Detection and Device ID Verification
Integrated telecom-level SIM swap checking to detect recently swapped SIM cards (a common fraud vector in banking), combined with device ID verification that matches the requesting device's identifier against the registered device binding to prevent cloned or spoofed device attacks.

## 4. Tri-Detector Face Detection Engine
Built a multi-detector ensemble combining:

- **Dlib HOG Detector** --- Fast frontal face detection with 68-point shape predictor for facial landmark extraction (jaw, brows, nose, eyes, mouth).
- **RetinaFace** --- State-of-the-art multi-scale pyramid network for detecting faces at difficult angles and scales.
- **MediaPipe Face Landmarker** --- 468 3D facial landmarks with 40+ blendshapes (facial action units) and Z-depth geometry information.
- Face matching via center-point Euclidean distance and IoU-based deduplication (threshold > 0.65).

## 5. 26+ Heuristic Artifact Signal Extractors
Developed a comprehensive heuristic engine detecting AI-generation artifacts including:

- Sharpness analysis (Laplacian variance)
- Noise uniformity detection (Gaussian difference, block-wise analysis)
- Frequency domain analysis (FFT high-frequency ratio, GAN grid artifacts)
- Color metrics (channel correlation, saturation variance, skin hue uniformity)
- Texture entropy (gradient magnitude distribution)
- Compression artifacts (DCT block boundary)
- Error Level Analysis (ELA --- JPEG recompression)
- L/R facial symmetry analysis
- 5 diffusion-model-specific signals (smooth skin, hue uniformity, low local contrast, compressed histogram, clean edges)
- 3D geometry validation (Z-depth flatness, frozen blink detection)
- 6 Dlib 68-point geometry checks (golden ratio deviation, jaw symmetry, landmark uniformity, nose bridge variance, brow level difference, EAR disagreement)

## 6. CLIP Multi-Prompt Zero-Shot Classification
Integrated OpenAI CLIP (ViT-Base-Patch32) with 5 prompt pairs for robust ensembling, bias correction (-0.10), outlier trimming, and probability calibration clamped to [0.05, 0.82].

## 7. Dedicated Deepfake Binary Classifier
Integrated `dima806/deepfake_vs_real_image_detection` from Hugging Face with temperature-scaled calibration (T=3.5) to prevent overconfidence, clamped to [0.05, 0.82].

## 8. Weighted Consensus Voting Engine
Built a three-model fusion system with weighted combination (Heuristic 20% + CLIP 45% + Deepfake 35%), consensus gating (requires 2 or more concerned signals to flag), hard overrides for strong agreement, and graceful fallback when models are unavailable.

## 9. Async Video Analysis Pipeline
Developed background video processing with configurable frame sampling (default every 30 frames), per-frame face detection and analysis, vote aggregation across frames, progress tracking, and in-memory job store with polling API.

## 10. EXIF and Metadata Intelligence
Built metadata scoring system that detects AI tool signatures (Stable Diffusion, Midjourney, DALL-E), screenshot detection (PNG + missing EXIF + common resolution), and ELA analysis for JPEG recompression artifacts.

## 11. Liveness Detection System
Implemented a state-machine-based liveness challenge system supporting blink detection (EAR threshold), mouth open detection (MAR threshold), and head nod detection --- preventing photo/video replay attacks.

## 12. Deepfake Detection Proxy Layer
Developed the Timble unified API gateway that proxies requests to three separate AI detection microservices (face at `localhost:8001`, voice at `localhost:8096`, document at `43.204.41.9`), supporting both synchronous and asynchronous workflows.

## 13. Auto-Migrating Database Schema
Designed PostgreSQL schema with four core tables (devices, auth\_sessions, auth\_context\_tokens, audit\_logs) with automatic migration on server startup.

## 14. In-Memory Session Orchestration
Built a concurrent-safe in-memory store for hybrid authentication session coordination with automatic cleanup every 60 seconds.

## 15. Device Lifecycle Management
Implemented full device management including registration, status checking, key rotation/update, revocation, and per-client per-user device binding with UPSERT logic.

## 16. Comprehensive Audit Logging
Built compliance-ready audit trail logging every authentication event with user reference, action, decision, IP address, device ID, and timestamp.

## 17. DEV\_MODE Simulation
Created a development mode that simulates device signing locally using a configurable private key, enabling full end-to-end testing without physical mobile devices.

## 18. API Documentation
Authored complete OpenAPI 3.0 specification, Swagger UI, integration guides, and quick reference documentation.

\newpage

# Brief Working of the Solution and Process Flow

## Overall Architecture

The Timble API server (`3.108.166.176:8097`) acts as a central authentication and fraud detection gateway. It sits between the bank's mobile app / backend server and the underlying verification services (PostgreSQL database, telecom SIM verification API, and AI detection microservices).

```
+--------------------------------------------------------------+
|                    Bank / Client Layer                        |
|   +------------------+           +---------------------+     |
|   |  Mobile App      |           |  Bank Backend       |     |
|   |  (Device Signing |           |  Server             |     |
|   |   + Camera)      |           |                     |     |
|   +--------+---------+           +----------+----------+     |
+------------|---------------------------------|---------------+
             | HTTPS                           | HTTPS
             +--------------+------------------+
                            |
+---------------------------+----------------------------------+
|                    Nginx Reverse Proxy                        |
|               (TLS Termination :443/:80)                     |
+---------------------------+----------------------------------+
                            | HTTP :8097
+---------------------------+----------------------------------+
|                    Timble API (Go)                            |
|          Authentication + Deepfake Gateway                   |
|   +--------------+--------------+--------------------+       |
|   |  Device Auth |  SIM Auth    |  Deepfake Proxy    |       |
|   |  (ECDSA)     |  (Swap Check)|  (Face/Voice/Doc)  |       |
|   +------+-------+------+-------+---------+----------+       |
+----------|--------------|-----------------|------------------+
           |              |                 |
     +-----+       +-----+         +-------+--------+
     |              |               |       |        |
+----v-----+ +-----v------+ +------v--+ +--v---+ +--v-------+
|PostgreSQL | | Telecom    | |Face Det.| |Voice | |Deepfake  |
| Database  | | SIM API    | |:8001    | |:8096 | |43.204.   |
| :5432     | | (SIM Swap) | |(Python) | |      | |41.9      |
+-----------+ +------------+ +---------+ +------+ +----------+
```

## Device Authentication Flow

1. **Registration** --- The mobile app generates an ECDSA P-256 key pair in the device's secure enclave (Android Keystore / iOS Secure Enclave). The public key is sent to Timble via `POST /v1/device/register` along with `client_id`, `user_ref`, `device_id`, `platform`, and `device_model`. Timble stores the device binding in PostgreSQL.

2. **Challenge Request** --- When the user needs to authenticate, the bank backend calls `POST /v1/auth/start` with `mode: "device"`. Timble generates a cryptographically secure 32-byte random challenge, stores it in the `auth_sessions` table with a 120-second TTL, and returns the challenge to the mobile app.

3. **Challenge Signing** --- The mobile app signs the challenge using the private key stored in the secure enclave (ECDSA SHA-256 signature) and submits the signature via `POST /v1/auth/complete`.

4. **Verification** --- Timble retrieves the stored public key for the device, verifies the ECDSA signature against the challenge, and if valid, generates an `auth_context_token` stored in the `auth_context_tokens` table with a 300-second TTL.

5. **Token Validation** --- The bank backend validates the token via `POST /v1/auth/verify` before granting the user access to sensitive operations.

## SIM Authentication Flow (with SIM Swap Check and Device ID Verification)

1. The bank backend calls `POST /v1/sim/start` with the user's phone number (MSISDN).

2. Timble authenticates with the telecom verification platform and initiates a **SIM Swap Check** --- verifying whether the SIM card linked to the phone number has been recently swapped, detecting potential SIM hijacking fraud.

3. Simultaneously, **Device ID Verification** is performed --- the device identifier (IMEI/device fingerprint) is matched against the registered device binding in the database to ensure the request is originating from the user's trusted device, not a cloned or spoofed device.

4. A `session_uri` is returned --- the user opens this URL in their mobile browser.

5. The telecom provider authenticates the user via the mobile network (SIM challenge over cellular data).

6. The bank backend polls `POST /v1/sim/complete` until a final verdict is received (ALLOW/DENY).

7. Authentication is approved **only if**: SIM swap is not detected AND device ID matches the registered binding.

## Hybrid Authentication Flow

1. The bank backend calls `POST /v1/auth/start` with `mode: "hybrid"`.

2. Timble creates both a device challenge AND a SIM session simultaneously.

3. The mobile app signs the device challenge AND the user opens the SIM redirect URL in parallel.

4. The backend calls `POST /v1/auth/complete` with the device signature.

5. Timble verifies **all factors** --- device signature must be valid AND SIM swap check must pass AND device ID must match --- before issuing the auth token.

## Face Detection and Deepfake Analysis Flow

### Image Analysis (Synchronous --- POST /v1/face/image)

**Stage 0: Dual-Model Pre-Check**

- Run CLIP on the full image (zero-shot AI vs Real classification using 5 prompt pairs).
- Run the Deepfake binary classifier on the full image.
- If **BOTH** models agree with 85% or higher confidence that the image is AI-generated, return early exit with verdict, skipping deeper analysis.

**Stage 1: Metadata and Physical Analysis**

- Extract EXIF metadata: camera make, model, lens, software, focal length.
- Score metadata for AI tool signatures (Stable Diffusion, Midjourney, DALL-E detection).
- Run Error Level Analysis (ELA): JPEG recompression artifact detection.
- Detect screenshots: PNG format + missing EXIF + common screen resolution + screen capture software tags.

**Stage 2: Tri-Detector Face Detection**

- **Dlib HOG Detector** --- Fast frontal face detection, returns bounding boxes, applies 68-point shape predictor extracting jaw (17 pts), brows (10 pts), nose (9 pts), eyes (12 pts), mouth (20 pts).
- **RetinaFace** --- Multi-scale pyramid network for difficult angles and occlusions.
- **MediaPipe Face Landmarker** --- 468 3D facial landmarks with 40+ blendshapes and Z-depth geometry.
- Face matching via center-point Euclidean distance, deduplication via IoU > 0.65.
- If 0 faces detected, return `NO_FACE` verdict.

**Stage 3: Per-Face Full Analysis (for each detected face)**

*Heuristic Analysis (20% weight):*

- **Sharpness** --- Laplacian variance: flag if >800 (over-sharp) or <30 (blurred)
- **Noise** --- Gaussian difference std dev: flag if <3.0 (too clean); spatial uniformity: flag if block-wise noise std <0.5
- **Frequency Domain (FFT)** --- High-frequency ratio: flag if >0.82; GAN grid artifacts: flag if max FFT/mean >50
- **Color Metrics** --- Channel correlation (R/G, R/B): flag if >0.97; saturation variance: flag if <25; skin hue variance: flag if <2.0
- **Texture Entropy** --- Gradient magnitude: flag if <1.5
- **Compression** --- DCT block boundary detection
- **ELA** --- JPEG recompression: flag if std <0.4
- **Symmetry** --- L/R pixel symmetry: flag if diff <3.0
- **Diffusion-Specific (5 signals)** --- Smooth skin (blur ratio >0.85), hue uniformity (std <2.0), low local contrast (block std <4.0), compressed histogram (<0.1% extremes), clean edges (Canny ratio 4-6.5%)
- **3D Geometry** --- Z-depth flatness (variance <0.00005), frozen blink (both eyes <0.02)
- **Dlib 68-Point Geometry (6 checks)** --- Golden ratio deviation (<0.05), jaw symmetry (<2.0), landmark uniformity (CV <0.48), nose bridge straightness (X-variance <1.5), brow level difference (<1.5), EAR disagreement (>0.15)

*CLIP Classification (45% weight):*

- 5 prompt pairs (real vs AI descriptions), forward pass per pair
- Trim high/low outliers, average middle scores
- Bias correction: subtract 0.10
- Calibration clamped to [0.05, 0.82]

*Deepfake Classifier (35% weight):*

- Binary classification on face crop (224x224)
- Temperature scaling (T=3.5) to prevent overconfidence
- Calibration clamped to [0.05, 0.82]

**Stage 4: Weighted Consensus Voting**

$$\text{final\_score} = (\text{heuristic} \times 0.20) + (\text{clip} \times 0.45) + (\text{deepfake} \times 0.35)$$

- **Consensus gate**: Requires 2 or more of 3 models to be "concerned" before flagging
- **Hard overrides**: Both models >0.82 forces score to 0.82 minimum; both <0.25 caps score at 0.30 maximum
- **Verdict**: AI\_GENERATED (score >= 0.65) | SUSPICIOUS (0.42--0.65) | REAL (< 0.42)
- Multiple faces use worst-case verdict aggregation

### Video Analysis (Asynchronous --- POST /v1/face/video)

1. Submit video file and receive a `job_id` and `poll_url`.
2. Background processing: strip audio, sample every Nth frame (default 30).
3. Per sampled frame: run full tri-detector + 3-model analysis pipeline.
4. Collect votes per frame: AI\_GENERATED / SUSPICIOUS / REAL.
5. Poll `GET /v1/face/video/{job_id}` for progress and results.
6. Final verdict from vote aggregation:
   - AI\_GENERATED in 40% or more frames results in AI\_GENERATED
   - AI\_GENERATED + SUSPICIOUS in 50% or more frames results in SUSPICIOUS
   - Otherwise REAL

### Liveness Detection

State-machine-based challenge system:

- **Blink Detection** --- Eye Aspect Ratio (EAR) threshold monitoring across consecutive frames
- **Mouth Open** --- Mouth Aspect Ratio (MAR) threshold detection
- **Head Nod** --- Head pose amplitude tracking via SolvePnP + RQDecomp3x3
- Prevents photo/video replay attacks by requiring real-time user response

### Voice Analysis (Asynchronous --- POST /v1/voice/analyze)

- Submit audio file and receive a `task_id`.
- AI analysis for voice cloning, synthetic speech, and speaker verification.
- Poll `GET /v1/voice/analyze/{task_id}` until completion.
- Returns verdict: authentic / synthetic / inconclusive.

### Document Analysis (Asynchronous --- POST /v1/deepfake/analyze)

- Submit document image and receive a `task_id`.
- AI analysis for document tampering, manipulation, and forgery detection.
- Poll `GET /v1/deepfake/analyze/{task_id}` until completion.
- Returns verdict: authentic / manipulated / inconclusive.

## Audit and Compliance

Every authentication attempt (success or failure) is recorded in the `audit_logs` table with: user reference, action type, decision, client IP address, device ID, and timestamp --- providing a complete compliance trail for regulatory requirements.

\newpage

# Feasibility of Implementation in Live Environment

1. **Already Deployed and Running** --- The solution is currently deployed on a live Linux server at `3.108.166.176:8097` with active telecom integration and AI detection services, proving real-world operational feasibility.

2. **Minimal Infrastructure Requirements** --- Requires only a single Go binary (auth API), Python FastAPI service (face detection), PostgreSQL database, and the detection microservices. No heavy application servers, no JVM, no container orchestration required for initial deployment.

3. **Distributed AI Processing** --- Compute-heavy deepfake/document detection runs on a separate server (`43.204.41.9`), face detection runs co-located at `localhost:8001`, and voice analysis at `localhost:8096` --- keeping the authentication API lightweight and responsive while AI workloads are isolated.

4. **Production Security** --- ECDSA P-256 cryptography meets banking-grade security standards. Configurable session TTLs (120s for challenges, 300s for tokens), API key authentication, and full audit logging support regulatory compliance (RBI, PSD2).

5. **Graceful Model Fallback** --- The face detection system works with partial model availability. If CLIP or the deepfake classifier is unavailable, the system falls back to the remaining models with adjusted weights, ensuring uninterrupted service.

6. **Nginx-Ready** --- Designed for Nginx reverse proxy deployment with TLS termination on ports 443/80, providing HTTPS encryption, rate limiting, and load balancing at the infrastructure layer.

7. **Zero-Downtime Deployments** --- Auto-migrating database schema ensures new versions can be deployed without manual migration steps or downtime.

8. **Telecom Integration Proven** --- SIM swap detection and device ID verification are fully functional with OAuth token refresh, session management, and polling --- tested against the live telecom verification API. SIM swap checks add a critical fraud prevention layer that detects hijacked phone numbers before authentication is granted.

9. **Environment-Based Configuration** --- All secrets, URLs, thresholds, model weights, and operational parameters are externalized via `.env` files, enabling seamless migration between development, staging, and production environments without code changes.

10. **Pre-Check Optimization** --- The dual-model pre-check (CLIP + Deepfake classifier at 85% threshold) enables early exit for obvious AI-generated images, reducing processing time and compute costs for clear-cut cases.

\newpage

# Technology Architecture

## System Architecture Overview

```
+--------------------------------------------------------------+
|                    Bank / Client Layer                        |
|   +------------------+           +---------------------+     |
|   |  Mobile App      |           |  Bank Backend       |     |
|   |  (Device Signing |           |  Server             |     |
|   |   + Camera)      |           |                     |     |
|   +--------+---------+           +----------+----------+     |
+------------|---------------------------------|---------------+
             | HTTPS                           | HTTPS
             +--------------+------------------+
                            |
+---------------------------+----------------------------------+
|                    Nginx Reverse Proxy                        |
|               (TLS Termination :443/:80)                     |
+---------------------------+----------------------------------+
                            | HTTP :8097
+---------------------------+----------------------------------+
|                    Timble API Server (Go 1.22)                |
|          Authentication + Deepfake Detection Gateway         |
|                                                              |
|   +------------------+  +---------------+  +---------------+ |
|   | Device Auth      |  | SIM Auth      |  | Deepfake      | |
|   | (ECDSA P-256)    |  | (SIM Swap +   |  | Detection     | |
|   |                  |  |  Device ID)   |  | Proxy Layer   | |
|   +--------+---------+  +-------+-------+  +---+---+---+---+ |
+------------|-------------------|----------------|---|---|-----+
             |                   |                |   |   |
       +-----+             +----+          +-----+   |   +------+
       |                   |               |         |          |
+------v------+  +---------v--------+  +---v---+  +-v----+  +--v-------+
| PostgreSQL  |  | Telecom SIM API  |  | Face  |  |Voice |  |Document  |
| Database    |  | (Sekura/XConnect)|  | Det.  |  |Det.  |  |Tampering |
| :5432       |  | SIM Swap Check + |  | :8001 |  |:8096 |  |43.204.   |
|             |  | Device ID Verify |  |       |  |      |  |41.9      |
+-------------+  +------------------+  +-------+  +------+  +----------+
```

## Backend Architecture (Go 1.22)

| Component | Technology | Purpose |
|:----------|:-----------|:--------|
| HTTP Server | Go stdlib `net/http`, `ServeMux` | Zero-framework API server on port 8097 |
| Database | PostgreSQL via `lib/pq` | Device bindings, auth sessions, tokens, audit logs |
| Cryptography | Go `crypto/ecdsa`, `crypto/elliptic` | ECDSA P-256 key generation, challenge signing, signature verification |
| Telecom Integration | HTTP client to Sekura API | OAuth token management, SIM swap check, device ID verification |
| Deepfake Proxy | HTTP clients to 3 microservices | Proxies face, voice, and document detection requests |
| Session Orchestration | In-memory store with mutex | Hybrid auth coordination with 60-second auto-cleanup |
| Database Migrations | Embedded SQL, auto-applied on startup | Zero-downtime schema evolution |

**Directory Structure:**

```
Fakeapp/
  cmd/api/              -> HTTP server entry point
  cmd/devsigner/        -> CLI tool for local challenge signing (dev mode)
  internal/
    config/             -> Environment variable configuration
    crypto/             -> ECDSA P-256 key operations
    handlers/           -> HTTP endpoint handlers
    models/             -> Domain entities (device, auth session)
    repository/         -> PostgreSQL query layer
    service/            -> Business logic (device, auth)
    deepfake/           -> HTTP clients for ML microservices
    sim/                -> Telecom provider integration (Sekura)
    orchestration/      -> Hybrid auth session coordination
    middleware/         -> CORS middleware
  migrations/           -> PostgreSQL schema files (auto-applied)
  static/               -> Single-page web UI (index.html)
  docs/                 -> API documentation, OpenAPI spec
```

## AI/ML Architecture (Python FastAPI)

| Component | Technology | Purpose |
|:----------|:-----------|:--------|
| API Framework | FastAPI 0.110+, Uvicorn | Async ML inference server on port 8001 |
| ML Framework | PyTorch 2.1+, TensorFlow/Keras | Model inference for face and video detection |
| Vision Models | Dlib 19.24, RetinaFace, MediaPipe 0.10 | Tri-detector face detection ensemble |
| Classification | CLIP ViT-Base-Patch32, HuggingFace Transformers 4.38+ | Zero-shot and binary deepfake classification |
| Image Processing | OpenCV 4.8+, Pillow 10.0+, SciPy 1.11+ | Heuristic signal extraction (FFT, ELA, Laplacian) |
| Video Processing | FFmpeg, OpenCV VideoCapture | Frame extraction, audio stripping, temporal analysis |
| Video Model | Xception + LSTM (TensorFlow/Keras) | Temporal deepfake detection across video frames |

**ML Model Pipeline:**

```
Input Image/Video
       |
       v
+------+------+
| Stage 0:    |  CLIP + Deepfake on full image
| Pre-Check   |  -> Early exit if both agree (85%+ confidence)
+------+------+
       |
+------+------+
| Stage 1:    |  EXIF metadata, AI tool signatures,
| Metadata    |  ELA analysis, screenshot detection
+------+------+
       |
+------+------+
| Stage 2:    |  Dlib HOG (68 pts) + RetinaFace (multi-scale)
| Tri-Detector|  + MediaPipe (468 pts + blendshapes)
| Face Detect |  -> IoU deduplication (>0.65)
+------+------+
       |
+------+------+   Per-face analysis:
| Stage 3:    |   - 26+ heuristic signals (20% weight)
| Full        |   - CLIP 5-prompt ensemble (45% weight)
| Analysis    |   - Deepfake classifier (35% weight)
+------+------+
       |
+------+------+
| Stage 4:    |  Weighted consensus voting
| Verdict     |  Consensus gate + hard overrides
| Fusion      |  -> REAL | SUSPICIOUS | AI_GENERATED
+-------------+
```

## Database Schema

| Table | Purpose | Key Columns |
|:------|:--------|:------------|
| `device_bindings` | Registered device public keys | client_id, user_ref, device_id, public_key_pem, status |
| `auth_sessions` | Challenge-response tracking | device_binding_id, challenge, signature, expires_at |
| `auth_context_tokens` | Issued authentication tokens | token, device_binding_id, expires_at |
| `audit_logs` | Compliance audit trail | user_ref, action, decision, ip_address, device_id, timestamp |

## External Integration Points

| Service | Protocol | Purpose |
|:--------|:---------|:--------|
| Sekura/XConnect Telecom API | HTTPS + OAuth | SIM swap detection + device ID verification |
| Face Detection (localhost:8001) | HTTP | AI-powered face deepfake detection |
| Voice VARE (localhost:8096) | HTTP | Voice authenticity risk analysis |
| Document Tampering (43.204.41.9) | HTTP | Document forgery detection |
| PostgreSQL (localhost:5432) | TCP | Persistent storage for device bindings, sessions, audit |

\newpage

# Banking Use Cases

## 1. Video KYC Deepfake Prevention

**Scenario:** During digital customer onboarding, a fraudster submits an AI-generated face image or deepfake video instead of a genuine selfie to open a bank account under a stolen identity.

**Timble's Role:** The face detection pipeline analyzes the submitted image/video through the tri-detector ensemble, 26+ heuristic signals, CLIP classification, and deepfake classifier. The weighted consensus engine flags AI-generated content with high confidence, preventing synthetic identity fraud at the onboarding stage.

**API Flow:** `POST /v1/face/image` (selfie) or `POST /v1/face/video` (video KYC recording)

## 2. SIM Swap Fraud Prevention for High-Value Transactions

**Scenario:** An attacker ports a customer's phone number to a new SIM card (via social engineering at the telecom store), then intercepts OTP codes to authorize fund transfers or account changes.

**Timble's Role:** Before processing any high-value transaction, the bank calls Timble's SIM authentication to verify whether the customer's SIM card has been recently swapped. If a swap is detected, the transaction is blocked and the customer is alerted through an alternate channel.

**API Flow:** `POST /v1/auth/start` with `mode: "sim"` -> SIM swap check + device ID verification

## 3. Secure Mobile Banking Authentication

**Scenario:** A customer logs into their mobile banking app to check balances, transfer funds, or pay bills.

**Timble's Role:** The mobile app uses ECDSA P-256 challenge-response authentication via the device's secure enclave. The private key never leaves the device hardware, making it immune to phishing, credential stuffing, and man-in-the-middle attacks. For high-risk operations (large transfers, beneficiary changes), hybrid mode adds SIM verification as a second factor.

**API Flow:** `POST /v1/auth/start` with `mode: "device"` (standard) or `mode: "hybrid"` (high-risk)

## 4. Voice Banking Fraud Prevention

**Scenario:** An attacker uses AI voice cloning technology to impersonate a customer during a phone banking call, attempting to authorize transactions or change account details via IVR or agent-assisted channels.

**Timble's Role:** The VARE (Voice Authenticity Risk Engine) analyzes the caller's audio in real time, detecting synthetic speech, voice cloning artifacts, and replay attacks. The ensemble scoring across multiple models (RawNet, SincNet, classifier, LLM reasoning) provides a risk-level verdict.

**API Flow:** `POST /v1/voice/analyze?sync=true` with base64-encoded audio

## 5. Document Forgery Detection for Loan Applications

**Scenario:** A fraudster submits digitally tampered salary slips, bank statements, or identity documents (Aadhaar, PAN, passport) to obtain a loan under false pretenses.

**Timble's Role:** The document tampering detection service analyzes submitted documents for manipulation artifacts, metadata inconsistencies, and pixel-level forgery indicators. Tampered documents are flagged before the loan is processed.

**API Flow:** `POST /v1/deepfake/analyze` with base64-encoded document image

## 6. Account Takeover Prevention

**Scenario:** An attacker gains partial access to a customer's account (through data breaches, phishing, or social engineering) and attempts to change recovery details, add new beneficiaries, or initiate transfers.

**Timble's Role:** Hybrid authentication (device + SIM) ensures that even if credentials are compromised, the attacker cannot authenticate without: (a) the customer's physical device with the cryptographic key in the secure enclave, AND (b) the customer's active SIM card passing the swap check. Both factors must pass simultaneously.

**API Flow:** `POST /v1/auth/start` with `mode: "hybrid"` -> parallel device challenge + SIM verification

## 7. Insurance Claim Fraud Detection

**Scenario:** A fraudster submits AI-generated photos of fabricated damage (vehicle accidents, property damage, medical reports) or forged supporting documents to file fraudulent insurance claims.

**Timble's Role:** Face detection identifies AI-generated photos, document tampering detection flags forged medical reports and repair estimates, and voice detection can verify the claimant's identity during recorded statements.

**API Flow:** Combination of `POST /v1/face/image`, `POST /v1/deepfake/analyze`, and `POST /v1/voice/analyze`

## 8. Regulatory Compliance (RBI/PSD2/FFIEC)

**Scenario:** Banks must comply with multi-factor authentication mandates from RBI (India), PSD2 Strong Customer Authentication (Europe), or FFIEC (US) requirements.

**Timble's Role:** A single platform provides all required authentication factors --- possession (device with cryptographic key), inherence (face liveness, voice biometrics), and network-level verification (SIM authentication) --- with a unified audit trail stored in PostgreSQL for regulatory reporting and forensic investigation.

**API Flow:** Full audit trail via `audit_logs` table with user_ref, action, decision, IP, device_id, timestamp

\newpage

# Future Roadmap

## Phase 1: Enhanced Detection Capabilities

| Feature | Description |
|:--------|:------------|
| **Lip-Sync Deepfake Detection** | Detect audio-visual mismatch in video calls where the lip movements don't match the spoken words --- a common artifact in real-time deepfake video attacks |
| **Age and Gender Verification** | Cross-reference detected face attributes against KYC records to detect identity mismatches during onboarding |
| **Multi-Language Voice Detection** | Extend VARE to support regional Indian languages (Hindi, Tamil, Telugu, Bengali) for phone banking fraud detection |
| **Real-Time Video Stream Analysis** | Analyze live video feeds during video KYC calls, not just recorded uploads, for instant deepfake detection |

## Phase 2: Platform Expansion

| Feature | Description |
|:--------|:------------|
| **FIDO2/WebAuthn Support** | Add passwordless authentication via FIDO2 hardware keys and platform authenticators alongside existing ECDSA device binding |
| **Biometric Template Matching** | On-device facial feature extraction and server-side template comparison for continuous authentication without storing raw biometric data |
| **Behavioral Biometrics** | Analyze typing patterns, touch pressure, swipe gestures, and device motion during mobile banking sessions for passive continuous authentication |
| **Multi-Bank SaaS Dashboard** | Web-based admin portal for banks to monitor authentication events, detection rates, fraud trends, and model performance across their customer base |

## Phase 3: Intelligence and Analytics

| Feature | Description |
|:--------|:------------|
| **Fraud Pattern Analytics** | ML-driven analysis of authentication and detection events to identify emerging fraud patterns, geographic hotspots, and attack campaigns before they scale |
| **Risk Scoring Engine** | Dynamic per-user risk scores based on authentication history, device changes, location anomalies, and detected fraud attempts --- enabling adaptive authentication that increases friction only for high-risk sessions |
| **Model Retraining Pipeline** | Automated feedback loop: flagged false positives/negatives from production are used to retrain and fine-tune detection models, keeping pace with evolving deepfake generators |
| **Threat Intelligence Feed** | Aggregate anonymized fraud signals across all bank clients to provide early warning of new attack techniques and deepfake generator signatures |

## Phase 4: Scale and Enterprise Readiness

| Feature | Description |
|:--------|:------------|
| **Kubernetes Deployment** | Containerized deployment with Helm charts, horizontal pod autoscaling, and GPU node pools for ML inference |
| **Multi-Region Architecture** | Regional API servers with geo-routed traffic, local PostgreSQL replicas, and centralized model serving for low-latency global authentication |
| **gRPC Internal Communication** | Replace HTTP-based inter-service communication with gRPC for reduced latency and schema enforcement between Go gateway and Python ML services |
| **Webhook and Event Streaming** | Real-time webhook notifications and Kafka-based event streaming for bank backends to receive authentication and detection events without polling |
| **SOC 2 Type II and ISO 27001** | Enterprise compliance certifications for security-sensitive banking deployments |

\newpage

# Market Potential and Scalability Aspects of the Solution

## Market Potential

1. **Target Market** --- Banks, NBFCs (Non-Banking Financial Companies), fintech platforms, payment gateways, insurance companies, and any organization requiring strong identity verification and fraud prevention.

2. **Growing Deepfake Threat** --- With the rapid advancement of generative AI, deepfake-based fraud (synthetic identity, voice cloning, document forgery) is becoming a critical threat to financial institutions. The global deepfake detection market is projected to grow significantly, making combined auth + deepfake detection a high-demand solution.

3. **Regulatory Tailwinds** --- Increasing regulatory requirements for multi-factor authentication (RBI mandates in India, PSD2/SCA in Europe, FFIEC in the US) create a compliance-driven demand for solutions that offer multiple authentication factors in a single platform.

4. **Unique Value Proposition** --- Unlike standalone authentication or standalone deepfake detection products, Timble combines both into a unified API, reducing integration complexity for banks and providing a single vendor for identity assurance.

5. **India-Specific Opportunity** --- With SIM verification integration already built for Indian telecoms and deployment on Indian infrastructure, the solution is positioned for the rapidly growing Indian digital banking market (UPI, digital lending, e-KYC).

6. **SIM Swap Fraud Prevention** --- SIM swap attacks are one of the fastest-growing fraud vectors in banking (account takeover via ported phone numbers). Built-in SIM swap detection addresses a critical pain point that banks are actively seeking solutions for, especially with the rise of OTP-based transactions in India.

7. **Multi-Modal Fraud Detection** --- The combination of face liveness, voice authenticity, and document integrity checks in a single platform addresses the full spectrum of identity fraud vectors --- from presentation attacks (photos/videos held up to cameras) to sophisticated AI-generated deepfakes.

## Scalability Aspects

1. **Horizontal API Scaling** --- The Go API server is stateless (session state in PostgreSQL, orchestration state in-memory with short TTL). Multiple instances can run behind Nginx load balancer for horizontal scaling.

2. **Independent AI Scaling** --- Detection microservices (face, voice, document) run on separate servers and can be independently scaled based on demand. GPU-heavy deepfake detection at `43.204.41.9` can be scaled separately from the lightweight auth API. The face detection service at `localhost:8001` can be replicated behind a load balancer for high-throughput image/video analysis.

3. **Database Scaling** --- PostgreSQL supports read replicas for scaling verification/read queries. The schema is designed with proper indexing (user\_ref, client\_id, device\_binding\_id) for efficient lookups at scale.

4. **Multi-Tenant Architecture** --- `client_id` based device isolation already supports multiple bank clients on a single deployment, enabling SaaS-model distribution without code changes.

5. **Microservice Modularity** --- New authentication modes (biometric, OTP, FIDO2) or new detection capabilities (lip-sync analysis, age verification) can be added as independent microservices without modifying the core platform.

6. **Low Resource Footprint** --- Go's compiled binary with zero-framework approach consumes minimal memory and CPU. The Python face detection service uses efficient model loading with HuggingFace caching, avoiding redundant downloads.

7. **Configurable Analysis Depth** --- Video frame sampling rate (`sample_every` parameter) allows trading accuracy for speed. The dual-model pre-check enables early exit for obvious cases, reducing compute costs at scale.

8. **Geographic Distribution** --- The architecture supports deploying API servers in multiple regions with a central database or region-specific databases, enabling low-latency authentication for geographically distributed users.
