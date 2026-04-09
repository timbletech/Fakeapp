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
