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
{\Large\textbf{SIM Swap Detection \& SIM Authentication}}\\[6pt]
{\large Individual Solution Overview}
\end{center}

\vspace{12pt}

\renewcommand{\arraystretch}{1.8}

\begin{longtable}{|>{\columncolor{sectionbg}\color{sectionfg}\bfseries\raggedright\arraybackslash}p{4.2cm}|>{\raggedright\arraybackslash}p{11cm}|}
\hline

Product Identity &
SIM Swap Detection is a telecom-level fraud prevention module that verifies whether a customer's SIM card has been recently swapped and whether the requesting device's identifier matches the registered device binding. It integrates with the Sekura/XConnect telecom verification API to perform real-time SIM swap checks and device ID verification over the mobile network, delivering an ALLOW or DENY verdict before any OTP or authentication is granted. The module operates as part of the Timble unified API gateway, accessible via \texttt{POST /v1/auth/start} with \texttt{mode: "sim"} or as part of hybrid authentication. \\
\hline

Problem Statement &
SIM swap fraud is one of the fastest-growing attack vectors in Indian banking. Attackers social-engineer telecom stores to port a victim's phone number to a new SIM card, then intercept OTP codes to authorize fund transfers, change recovery details, and take over accounts. OTP-based authentication --- the most widely used second factor in Indian banking --- provides zero protection against ported numbers. Banks have no visibility into whether a SIM has been recently swapped before sending an OTP. Losses from SIM swap fraud run into thousands of crores annually, and the attack is invisible to traditional authentication systems. \\
\hline

What It Demonstrates &
Real-time telecom-level fraud detection that catches SIM swap attacks before authentication is granted: (1) Direct integration with the Sekura/XConnect telecom verification platform via OAuth-authenticated API calls (Basic Auth for token refresh, Bearer token for operations); (2) SIM swap check --- queries the telecom network to determine whether the SIM card linked to the customer's MSISDN (phone number) has been recently ported or swapped, with a configurable lookback window; (3) Device ID verification --- matches the requesting device's identifier (IMEI/device fingerprint) against the registered device binding in the database, catching cloned or spoofed devices; (4) Mobile network authentication --- redirects the user to the telecom operator's portal for network-level SIM challenge over cellular data, ensuring the phone number is authenticated by the carrier, not just the application. Both checks must pass for ALLOW verdict. \\
\hline

Technology Architecture &
Go API gateway (port 8097) integrates with Sekura/XConnect telecom API via HTTP client with 30-second timeout. Authentication flow: (1) \texttt{POST /v1/auth/start} (mode: ``sim'') --- Timble authenticates with Sekura via \texttt{POST /v1/token} using Basic Auth (client\_key:client\_secret) with refresh\_token grant, obtaining a Bearer access token; (2) \texttt{POST /v1/insights/\{msisdn\}} --- sends the customer's phone number for SIM swap check and device ID verification against registered binding; (3) returns a \texttt{session\_uri} --- the user opens this URL in their mobile browser for carrier-level SIM challenge over cellular data; (4) \texttt{GET /v1/sim/poll/\{session\_id\}} --- bank backend polls until the telecom operator returns a final verdict (ALLOW/DENY); (5) result is stored in the in-memory session store and audit logged. Configuration: Sekura base URL, client key, client secret, and refresh token are externalized via environment variables. Session state is tracked in an in-memory store with automatic cleanup. Full audit trail in PostgreSQL (user\_ref, action, decision, IP, device\_id, timestamp). \\
\hline

Banking Use Cases &
High-Value Transaction Authorization --- verify SIM integrity before processing large fund transfers, beneficiary additions, or NEFT/RTGS/IMPS transactions above threshold amounts, blocking transactions if a SIM swap is detected. Account Takeover Prevention --- detect SIM hijacking before attackers can intercept OTPs to change passwords, recovery emails, or registered mobile numbers. New Account Fraud --- verify the applicant's SIM during digital account opening to ensure the phone number hasn't been recently ported (a common indicator of synthetic identity fraud). Hybrid Authentication --- combine SIM verification with ECDSA device challenge in parallel for high-risk operations, requiring both the physical device AND the verified SIM to pass before granting access. Regulatory Compliance --- provide auditable SIM verification records for RBI's authentication guidelines and PSD2 Strong Customer Authentication requirements. \\
\hline

Future Roadmap &
Near-term: Multi-carrier support across all major Indian telecom operators (Jio, Airtel, Vi, BSNL), configurable SIM swap lookback window (24hr, 48hr, 7-day), real-time webhook notifications on SIM swap events. Mid-term: SIM swap risk scoring based on historical patterns (frequency of swaps, geographic anomalies, time-of-day signals), integration with TRAI's SIM swap notification framework. Long-term: Cross-carrier fraud intelligence sharing (anonymized swap patterns), eSIM transition detection as eSIM adoption grows, and international roaming-aware verification for NRI banking customers. \\
\hline

Market Potential and Scalability Aspects of the Solution &
India has 1.15B+ mobile subscribers with SIM-linked banking. SIM swap fraud losses are estimated at Rs 1,000+ crores annually. RBI's increasing focus on authentication security and the mandate for multi-factor verification for digital transactions creates a regulatory-driven market. The SIM swap detection module scales with the stateless Go API gateway --- multiple instances behind Nginx load balancer, with each instance making independent Sekura API calls. OAuth token management is per-request (no shared cache), enabling clean horizontal scaling. The in-memory session store supports thousands of concurrent verification sessions with automatic 60-second cleanup. As India's digital payment ecosystem grows (10B+ monthly UPI transactions), every high-value transaction is a potential verification request, representing a massive per-call revenue opportunity. \\
\hline

\end{longtable}
