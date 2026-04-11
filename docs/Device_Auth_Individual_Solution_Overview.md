---
geometry: margin=2.5cm
fontsize: 11pt
header-includes:
  - \usepackage{booktabs}
  - \usepackage{longtable}
  - \usepackage{array}
  - \usepackage{colortbl}
  - \usepackage{xcolor}
  - \definecolor{sectionbg}{HTML}{1B3A5C}
  - \definecolor{sectionfg}{HTML}{FFFFFF}
  - \pagestyle{empty}
---

\begin{center}
{\Large\textbf{Device-Based Cryptographic Authentication}}\\[6pt]
{\large Individual Solution Overview}
\end{center}

\vspace{12pt}

\renewcommand{\arraystretch}{1.8}

\begin{longtable}{|>{\columncolor{sectionbg}\color{sectionfg}\bfseries\raggedright\arraybackslash}p{4.2cm}|>{\raggedright\arraybackslash}p{11cm}|}
\hline

Product Identity &
Device Authentication is a cryptographic challenge-response authentication system built on ECDSA P-256 (secp256r1) elliptic curve cryptography. The mobile app generates a key pair in the device's secure enclave (Android Keystore / iOS Secure Enclave), registers the public key with the server, and authenticates by signing 32-byte random challenges. The private key never leaves the device hardware, making it immune to phishing, credential stuffing, man-in-the-middle attacks, and server-side breaches. The system manages full device lifecycle: registration, authentication, key rotation, status checking, and revocation, with per-client per-user device binding and a compliance-grade audit trail. \\
\hline

Problem Statement &
Password and OTP-based authentication are fundamentally broken for banking security. Passwords are phished, leaked in breaches, reused across services, and vulnerable to credential stuffing at scale. SMS OTPs are intercepted via SIM swap attacks, SS7 vulnerabilities, and malware. TOTP apps require user action and are susceptible to real-time phishing proxies (EvilGinx-style attacks). Banks need an authentication factor that cannot be phished, intercepted, or replayed --- one tied to a specific physical device with cryptographic proof of possession, where compromise of the server database does not compromise authentication credentials. \\
\hline

What It Demonstrates &
Banking-grade cryptographic device authentication without passwords or OTPs: (1) ECDSA P-256 key pair generation --- the mobile app creates a 256-bit elliptic curve key pair in the device's hardware-backed secure enclave, registering only the public key (PEM format) with the server via \texttt{POST /v1/device/register}; (2) challenge-response protocol --- server generates a cryptographically secure 32-byte random challenge (base64-encoded) with a configurable 120-second TTL, stored in PostgreSQL; (3) device signing --- the mobile app signs the challenge using SHA-256 hashing + ECDSA signing in the secure enclave, producing an ASN.1-encoded (r, s) signature; (4) server verification --- the server retrieves the stored public key, verifies the ECDSA signature against the challenge, and issues an auth\_context\_token (300-second TTL) on success; (5) full device lifecycle --- register, authenticate, check status, update keys, revoke binding, with UPSERT logic for re-registration; (6) DEV\_MODE simulation for end-to-end testing without physical mobile devices. \\
\hline

Technology Architecture &
Go 1.22 API server using stdlib \texttt{net/http} and \texttt{crypto/ecdsa} (zero external crypto dependencies). Endpoints: \texttt{POST /v1/device/register} (register public key + device info), \texttt{POST /v1/auth/start} mode ``device'' (generate 32-byte challenge), \texttt{POST /v1/auth/complete} mode ``device'' (verify ECDSA signature), \texttt{POST /v1/auth/verify} (validate auth\_context\_token), \texttt{GET /v1/device/check} (check binding status), \texttt{PUT /v1/device/update} (update device info / rotate key), \texttt{POST /v1/device/revoke} (revoke binding). PostgreSQL schema: \texttt{device\_bindings} table (client\_id, user\_ref, device\_id, public\_key\_pem, platform, device\_model, os\_version, status, created\_at), \texttt{auth\_sessions} table (challenge, signature, device\_binding\_id, expires\_at), \texttt{auth\_context\_tokens} table (token, device\_binding\_id, expires\_at), \texttt{audit\_logs} table (user\_ref, action, decision, ip\_address, device\_id, timestamp). Auto-migrating schema applied on server startup. Configurable TTLs via environment variables (CHALLENGE\_EXPIRY\_SECONDS=120, AUTH\_TOKEN\_EXPIRY\_SECONDS=300). \\
\hline

Banking Use Cases &
Secure Mobile Banking Login --- replace password + OTP with cryptographic device authentication that cannot be phished, intercepted, or replayed, providing a frictionless yet banking-grade login experience. High-Value Transaction Authorization --- step-up authentication for large transfers, beneficiary additions, and account changes using a fresh cryptographic challenge that expires in 120 seconds. Multi-Device Management --- per-client per-user device binding with status tracking, enabling banks to manage which devices are authorized per customer and revoke compromised devices instantly. Hybrid Authentication --- combine device cryptographic proof with SIM swap verification in parallel for the highest-risk operations, requiring both factors to pass before granting access. Compliance Audit --- every authentication attempt (success or failure) is recorded in the audit\_logs table with user\_ref, action, decision, client IP, device\_id, and timestamp --- providing a complete trail for RBI, PSD2/SCA, and FFIEC regulatory audits. \\
\hline

Future Roadmap &
Near-term: FIDO2/WebAuthn support to enable passwordless authentication via hardware security keys and platform authenticators alongside ECDSA device binding, client SDKs (iOS/Android) for faster bank integration. Mid-term: Continuous authentication via periodic silent re-challenge during active sessions, risk-adaptive challenge frequency based on transaction amount and user behavior patterns, multi-device authorization policies (allow N devices per user with admin controls). Long-term: Behavioral biometrics integration (typing cadence, touch pressure, device motion) as a passive continuous authentication layer alongside cryptographic challenges, cross-device attestation for secure device migration workflows, and hardware attestation verification (Android SafetyNet/Play Integrity, iOS DeviceCheck) to ensure keys are generated in genuine secure enclaves. \\
\hline

Market Potential and Scalability Aspects of the Solution &
India has 750M+ smartphone users with mobile banking adoption growing rapidly. RBI mandates for strong customer authentication and the deprecation of SMS OTP as a sole second factor create regulatory demand for device-based cryptographic authentication. The global passwordless authentication market is projected to reach \$25B+ by 2030. Device Authentication scales horizontally --- the Go API server is stateless (all state in PostgreSQL), enabling multiple instances behind Nginx load balancer. PostgreSQL supports read replicas for scaling verification queries. The schema is indexed on (client\_id, user\_ref, device\_id) for efficient lookups at scale. Challenge generation and signature verification are CPU-lightweight operations (sub-millisecond ECDSA verification), supporting thousands of concurrent authentications per instance. Multi-tenant client\_id isolation enables SaaS deployment serving multiple banks on a single infrastructure. \\
\hline

\end{longtable}
