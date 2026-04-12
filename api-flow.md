# Timble API Flow & Architecture

Diagrams showing how the **Bank (client)** integrates with the **Timble authentication server** across all supported auth modes.

---

## Deployment Architecture

```mermaid
graph TD
    subgraph Bank["🏦 Bank - Client"]
        BankApp["📱 Bank Mobile App"]
        BankSrv["🖥️ Bank Backend Server"]
    end
    
    subgraph Timble["🖥️ Timble Linux Server<br/>3.108.166.176"]
        Nginx["Nginx :443 / :80"]
        Go["Go API :8097"]
        PG["PostgreSQL :5432"]
        FS["📁 Static Files<br/>/opt/timble/static"]
        Migrations["🔄 Migrations<br/>/opt/timble/migrations"]
    end
    
    BankApp -->|HTTPS| Nginx
    BankSrv -->|HTTPS| Nginx
    Nginx -->|proxy_pass| Go
    Go -->|DB_DSN| PG
    Go --> FS
    Go -->|on startup| Migrations
```

---

## Flow 1 — Device Authentication

The bank mobile app holds an EC key pair. Timble issues a challenge; the app signs it; Timble verifies the signature. The bank backend validates the resulting token server-side.

```mermaid
sequenceDiagram
    participant BankApp as Bank Mobile App
    participant BankSrv as Bank Backend
    participant Timble as Timble API
    participant DB as PostgreSQL

    Note over BankApp,Timble: One-time Device Registration
    BankApp->>Timble: POST /v1/device/register
    Note right of Timble: client_id, user_ref, public_key, device_info
    Timble->>DB: Store device binding + public key
    Timble-->>BankApp: device_binding_id

    Note over BankApp,Timble: Auth Flow - every login
    BankApp->>Timble: POST /v1/auth/start
    Note right of Timble: mode=device, device_binding_id
    Timble->>DB: Create auth session + challenge
    Timble-->>BankApp: auth_session_id, challenge, challenge_id

    BankApp->>BankApp: Sign challenge with EC private key

    BankApp->>Timble: POST /v1/auth/complete
    Note right of Timble: auth_session_id, challenge_id, device_signature
    Timble->>DB: Verify signature against stored public key
    Timble-->>BankApp: decision=ALLOW, auth_context_token

    Note over BankSrv,Timble: Bank backend validates token server-side
    BankSrv->>Timble: POST /v1/auth/verify
    Note right of Timble: client_id, auth_context_token
    Timble-->>BankSrv: valid=true
```

---

## Flow 2 — SIM Authentication

The bank backend drives the API calls. The bank app is redirected to a Sekura URL where the SIM card challenge is resolved silently over the mobile network. The backend polls Timble until a final decision is returned.

```mermaid
sequenceDiagram
    participant BankApp as Bank Mobile App
    participant BankSrv as Bank Backend
    participant Timble as Timble API
    participant Sekura as Sekura XConnect

    BankSrv->>Timble: POST /v1/sim/start
    Note right of Timble: client_id, user_ref, msisdn + X-API-Key
    Timble->>Sekura: Create SIM session
    Sekura-->>Timble: session_uri
    Timble-->>BankSrv: auth_session_id, session_uri

    BankSrv-->>BankApp: Forward session_uri
    BankApp->>Sekura: Open session_uri in device browser
    Sekura->>Sekura: SIM challenge resolved via mobile network

    loop Poll until final
        BankSrv->>Timble: POST /v1/sim/complete
        Note right of Timble: auth_session_id + X-API-Key
        Timble->>Sekura: Check session status
        Sekura-->>Timble: status
        Timble-->>BankSrv: 202 PENDING or 200 ALLOW/DENY
    end
```

---

## Flow 3 — Hybrid Authentication

Combines both factors: device cryptography **and** SIM verification in a single session. The app signs the device challenge and opens the SIM redirect in parallel. The bank backend polls Timble until both factors are confirmed.

```mermaid
sequenceDiagram
    participant BankApp as Bank Mobile App
    participant BankSrv as Bank Backend
    participant Timble as Timble API
    participant DB as PostgreSQL
    participant Sekura as Sekura XConnect

    Note over BankApp,Timble: One-time device registration - same as Device Flow

    BankSrv->>Timble: POST /v1/auth/start
    Note right of Timble: mode=hybrid, device_binding_id, msisdn
    Timble->>DB: Create session + challenge
    Timble->>Sekura: Create SIM session
    Timble-->>BankSrv: auth_session_id, challenge, sim session_uri

    BankSrv-->>BankApp: Forward challenge + session_uri

    par Bank App signs device challenge
        BankApp->>BankApp: Sign challenge with EC private key
    and Bank App triggers SIM check
        BankApp->>Sekura: Open session_uri in device browser
        Sekura->>Sekura: SIM resolved via mobile network
    end

    BankSrv->>Timble: POST /v1/auth/complete
    Note right of Timble: auth_session_id, device_signature
    Timble->>DB: Verify device signature
    Timble->>Sekura: Check SIM result

    alt Both factors verified
        Timble-->>BankSrv: decision=ALLOW, auth_context_token
    else SIM still in progress
        Timble-->>BankSrv: 202 decision=PENDING, poll again
    end
```

---

## Flow 4 — New Device Verification (Unknown Device Approval)

When a user tries to authenticate from a new/unknown device, the system auto-triggers an approval request sent to their existing trusted device. The trusted device owner approves or denies the request. On approval, the new device is registered and can proceed with normal authentication.

```mermaid
sequenceDiagram
    participant NewDev as New Device
    participant Timble as Timble API
    participant DB as PostgreSQL
    participant MainDev as Trusted Device

    Note over NewDev,Timble: New device attempts login
    NewDev->>Timble: POST /v1/auth/start
    Note right of Timble: mode=device, device_id (unknown), public_key
    Timble->>DB: Check device binding
    DB-->>Timble: DEVICE_NOT_MATCH

    Note over Timble,DB: Auto-trigger approval flow
    Timble->>DB: Create device_approval_request (PENDING)
    Timble-->>NewDev: 202 next_step=DEVICE_APPROVAL_REQUIRED, approval_id

    Note over MainDev,Timble: Trusted device polls for pending requests
    loop Poll pending
        MainDev->>Timble: GET /v1/auth/device-verify/pending
        Note right of Timble: client_id, user_ref
        Timble->>DB: Fetch PENDING approvals
        Timble-->>MainDev: [{approval_id, requesting_device_info}]
    end

    MainDev->>Timble: POST /v1/auth/device-verify/respond
    Note right of Timble: approval_id, action=approve
    Timble->>DB: Update status to APPROVED
    Timble->>DB: Auto-register new device with public_key
    Timble-->>MainDev: status=APPROVED

    Note over NewDev,Timble: New device polls for result
    loop Poll status
        NewDev->>Timble: POST /v1/auth/device-verify/status
        Note right of Timble: client_id, approval_id
        Timble-->>NewDev: status=PENDING / APPROVED / DENIED
    end

    Note over NewDev,Timble: Device now trusted - normal auth
    NewDev->>Timble: POST /v1/auth/start
    Note right of Timble: mode=device (now recognized)
    Timble-->>NewDev: auth_session_id, challenge

    NewDev->>NewDev: Sign challenge with EC private key

    NewDev->>Timble: POST /v1/auth/complete
    Note right of Timble: auth_session_id, device_signature
    Timble-->>NewDev: decision=ALLOW, auth_context_token
```