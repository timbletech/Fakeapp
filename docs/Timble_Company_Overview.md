# Timble - Company Overview

---

## 1. Unique Selling Proposition (USP)

Timble is a **unified fraud-prevention and authentication API platform** that combines four critical security capabilities into a single, integrated solution:

- **Device-Based Cryptographic Authentication** — ECDSA P-256 challenge-response signatures ensuring tamper-proof device verification.
- **SIM / SIM-Swap Verification** — Real-time telecom-level fraud detection via Sekura/XConnect integration, identifying SIM swap attacks before they succeed.
- **Device Binding** — Persistent cryptographic binding of user identity to a specific physical device.
- **AI-Powered Deepfake Detection** — Multi-modal analysis covering face (image & video), voice, and document tampering.

**Strongest Differentiator:**
The **hybrid authentication flow** — device cryptographic verification and SIM verification execute simultaneously in parallel, delivering banking-grade security without sacrificing user experience. This is combined with multi-modal AI fraud checks purpose-built for regulated financial services.

---

## 2. Target Market

| Segment | Examples |
|---|---|
| Banks | Retail banks, corporate banks, digital-only banks |
| NBFCs | Non-Banking Financial Companies |
| Fintech Platforms | Payment apps, neobanks, lending platforms |
| Digital Lenders | Online lending, micro-lending, BNPL providers |
| Payment Gateways | Transaction processors, UPI service providers |
| Wallet Apps | Mobile wallets, prepaid instruments |
| Insurance Companies | Digital insurance platforms, insurtech |

**Primary Focus:** Indian BFSI market (RBI-regulated institutions)
**Expansion Markets:** PSD2/SCA-regulated entities (Europe), FFIEC-compliant institutions (United States)

**Key Use Cases:**
- Secure login and session management
- Transaction authorization and step-up authentication
- e-KYC and customer onboarding verification
- Ongoing anti-fraud checks across the customer lifecycle

---

## 3. Market Size

| Metric | Details |
|---|---|
| **Global Digital Identity Verification Market** | ~$10-12B (2024), projected ~$30-35B by 2030 |
| **Global Deepfake Detection Market** | Projected ~$5-10B by 2030 |
| **India Digital Banking Security & e-KYC** | Fast-growing subset driven by UPI adoption (10B+ monthly transactions) and RBI mandates |

**Market Drivers:**
- Rising deepfake sophistication making traditional checks insufficient
- Regulatory mandates (RBI digital lending guidelines, PSD2 Strong Customer Authentication, FFIEC authentication guidance)
- Explosive growth of digital banking and fintech in India
- SIM swap fraud emerging as a top attack vector in financial services

*Note: Detailed TAM/SAM/SOM to be calculated based on finalized pricing model and go-to-market scope.*

---

## 4. Business Model

**Model Type:** B2B SaaS / API-as-a-Service

**Revenue Streams:**

| Stream | Description |
|---|---|
| Enterprise Licensing | Platform access and integration fees per client |
| Usage-Based Pricing | Per-call charges for authentication, SIM verification, and deepfake analysis API requests |
| Integration & Setup | Professional services for enterprise onboarding and custom deployment |

**Architecture Supporting the Model:**
- **Multi-tenant design** — Client ID-based isolation supports multiple bank clients on a single deployment without code changes
- **API key authentication** — Secure, per-client access management
- **Modular microservice architecture** — Independent scaling of auth, SIM, and AI detection services

**Cost Structure:**
- Telecom verification fees (Sekura/XConnect per SIM check)
- AI/GPU inference costs (per deepfake analysis — face, voice, document)
- Cloud infrastructure and hosting
- Enterprise support and compliance maintenance

---

## 5. Traction

### Sales & Revenue
Currently in pre-revenue / early commercialization stage. No sales or customer revenue data available.

### Product Milestones Achieved

| Milestone | Status |
|---|---|
| Fully functional Go API server with three auth modes (device, SIM, hybrid) | Complete |
| AI deepfake detection — face image, face video, voice, document tampering | Complete |
| Synchronous and asynchronous (polling) workflows for all AI analysis | Complete |
| OpenAPI 3.0.3 specifications published | Complete |
| Bank mobile integration guide for enterprise onboarding | Complete |
| Demo UI for client evaluation and testing | Complete |
| End-to-end test suite covering all authentication and detection flows | Complete |
| PostgreSQL persistence with compliance-grade audit logging | Complete |
| Live Sekura/XConnect telecom integration for SIM swap detection | Complete |
| Production deployment on live Linux server with Nginx TLS termination | Complete |

### Technical Readiness
- **Production server:** Deployed and running (3.108.166.176:8097)
- **Audit trail:** Complete compliance logging — user, action, decision, IP address, device ID, timestamp
- **Security standard:** Banking-grade ECDSA P-256 cryptography
- **Integration readiness:** OpenAPI specs + dedicated bank integration documentation

---

## 6. Profit Margins

| Metric | Estimate |
|---|---|
| **Expected Gross Margin** | 70-85% (typical for SaaS/API platforms) |
| **Margin Moderators** | Telecom verification fees, AI inference compute costs, cloud hosting, enterprise support |
| **Margin Improvement Path** | Economies of scale — fixed platform costs amortized across growing client base and API call volume |

*Note: Detailed margin analysis to be developed once pricing and initial client contracts are finalized.*

---

## 7. Financial Projections

Not yet developed. To be built based on:

- Finalized pricing model (per-call rates, licensing tiers)
- Target client pipeline and conversion assumptions
- Projected API call volumes per client
- Go-to-market timeline and sales cycle estimates
- Infrastructure cost scaling projections

---

## 8. Current Funding

| Detail | Status |
|---|---|
| **Amount** | Not disclosed |
| **Round** | Not disclosed |
| **Date** | Not disclosed |

---

## 9. Planned Amount for Next Round

Not disclosed. To be determined based on:
- Go-to-market expansion requirements
- Enterprise sales team build-out
- AI model improvement and compute scaling needs
- Regulatory certification costs
- Multi-region infrastructure deployment

---

## 10. Company Roadmap

### Near-Term

| Priority | Description |
|---|---|
| Enterprise Client Onboarding | Expand bank client integrations and multi-client production deployments |
| Production Hardening | Horizontal scaling, load balancing, multi-region deployment, database read replicas |
| Compliance Strengthening | Deepen audit capabilities for RBI, PSD2/SCA, and FFIEC regulatory requirements |

### Mid-Term

| Priority | Description |
|---|---|
| AI Enhancement | Improve deepfake detection accuracy; add lip-sync analysis, age verification, liveness enhancements |
| New Auth Modes | Add biometric, OTP, and FIDO2 authentication as independent microservices |
| Client SDKs | Build SDKs and client libraries from OpenAPI specifications for faster integration |

### Long-Term

| Priority | Description |
|---|---|
| Enterprise Controls | Implement rate limiting, monitoring dashboards, and per-client policy management for SLA compliance |
| Full Lifecycle Coverage | Extend fraud stack across onboarding, login, transaction approval, and KYC workflows |
| Cost Optimization | Scale configurable analysis depth (video frame sampling) to optimize cost vs. accuracy per client |
| Geographic Expansion | Multi-region API deployment for latency-sensitive banking applications |

---

*Document generated from Timble codebase and project documentation.*
