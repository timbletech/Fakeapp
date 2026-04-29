package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"device_only/internal/sim/model"
	"device_only/internal/sim/provider"
	"device_only/internal/sim/store"

	"github.com/google/uuid"
)

// SessionError carries a machine-readable code, a human message, and an HTTP
// status so the handler can build a consistent error response without any
// business-logic knowledge.
type SessionError struct {
	Code    string
	Message string
	HTTP    int
}

func (e *SessionError) Error() string { return e.Message }

// AuthResult is the successful outcome of CompleteAuth. Exactly one of Complete
// or Pending is non-nil depending on whether the device has finished the SIM
// challenge.
type AuthResult struct {
	Complete  *model.CompleteResponse
	Pending   *model.PendingResponse
	IsPending bool
}

// AuthService orchestrates SIM verification by coordinating the Sekura provider,
// session store, and decision logic.
type AuthService struct {
	store       *store.SessionStore
	sekura      *provider.SekuraProvider
	baseURL     string
	sessionTTL  time.Duration
	maxAttempts int
	devMode     bool
}

// NewAuthService wires dependencies and returns a ready-to-use AuthService.
// When devMode is true, upstream Sekura failures are tolerated with mock data so
// developers can run the flow without the integration being live; in production
// (devMode=false) StartAuth fails closed with an upstream error instead.
func NewAuthService(
	store *store.SessionStore,
	sekura *provider.SekuraProvider,
	baseURL string,
	sessionTTLSeconds int,
	maxAttempts int,
	devMode bool,
) *AuthService {
	return &AuthService{
		store:       store,
		sekura:      sekura,
		baseURL:     strings.TrimRight(baseURL, "/"),
		sessionTTL:  time.Duration(sessionTTLSeconds) * time.Second,
		maxAttempts: maxAttempts,
		devMode:     devMode,
	}
}

// StartAuth validates the request, calls Sekura to obtain device-match URIs and
// the sim-swap-check result, persists a new session, and returns the data the
// bank app needs to trigger the SIM challenge on the device.
func (s *AuthService) StartAuth(req *model.StartRequest) (*model.StartResponse, error) {
	log.Printf("[INFO] StartAuth: client_id=%s user_ref=%s msisdn=%s action=%s",
		req.ClientID, req.UserRef, req.MSISDN, req.Action)

	token, err := s.sekura.GetToken()
	if err != nil {
		if !s.devMode {
			log.Printf("[ERROR] Sekura GetToken failed: %v", err)
			return nil, fmt.Errorf("sim_provider_unavailable: %w", err)
		}
		log.Printf("[WARN] Sekura GetToken failed in DEV_MODE, using mock token: %v", err)
		token = "mock_token"
	}

	insights, err := s.sekura.GetInsights(req.MSISDN, token)
	if err != nil {
		if !s.devMode {
			log.Printf("[ERROR] Sekura GetInsights failed: %v", err)
			return nil, fmt.Errorf("sim_provider_unavailable: %w", err)
		}
		log.Printf("[WARN] Sekura GetInsights failed in DEV_MODE, using mock insights: %v", err)
		insights = &model.SekuraInsightsResponse{}
		insights.DeviceMatch.SessionURI = "https://in.safr.xconnect.net/v1/dm/session/mock-session"
		insights.DeviceMatch.PollingURI = "https://in.safr.xconnect.net/v1/dm/polling/mock-session"
		insights.SimSwapCheck.Result = false
	}

	now := time.Now().UTC()
	sessionID := "sess_" + uuid.New().String()

	session := &model.Session{
		AuthSessionID: sessionID,
		ClientID:      req.ClientID,
		UserRef:       req.UserRef,
		MSISDN:        req.MSISDN,
		Action:        req.Action,
		SessionURI:    insights.DeviceMatch.SessionURI,
		PollingURI:    insights.DeviceMatch.PollingURI,
		SimSwapResult: insights.SimSwapCheck.Result,
		SimSwapDate:   insights.SimSwapCheck.Date,
		Status:        "pending",
		Decision:      "",
		Attempts:      0,
		CreatedAt:     now,
		ExpiresAt:     now.Add(s.sessionTTL),
	}

	s.store.Set(session)
	log.Printf("[INFO] Session created: session_id=%s expires_at=%s", sessionID, session.ExpiresAt.Format(time.RFC3339))

	wrappedURI := fmt.Sprintf("%s/v1/sim/redirect/%s", s.baseURL, sessionID)
	pollingBase := strings.Replace(s.baseURL, "http://", "https://", 1)
	wrappedPollingURI := fmt.Sprintf("%s/v1/sim/poll/%s", pollingBase, sessionID)

	return &model.StartResponse{
		AuthSessionID:  sessionID,
		ExpiresIn:      int(s.sessionTTL.Seconds()),
		NextStep:       "SIM_CHALLENGE_REQUIRED",
		SessionURI:     wrappedURI,
		PollingURI:     wrappedPollingURI,
		Instructions:   "Load session_uri on the mobile device over mobile data (not WiFi). The device must use cellular network for SIM verification to work.",
		SimSwapCheck:   &insights.SimSwapCheck,
		OperatorLookup: &insights.OperatorLookup,
	}, nil
}

// CompleteAuth looks up the session, polls Sekura for the device-match result,
// combines it with the stored sim-swap result, and returns either a pending
// notice or a final ALLOW/DENY decision.
func (s *AuthService) CompleteAuth(authSessionID string) (*AuthResult, *SessionError) {
	session, ok := s.store.Get(authSessionID)
	if !ok {
		return nil, &SessionError{
			Code:    "SESSION_NOT_FOUND",
			Message: "No session found with the provided auth_session_id.",
			HTTP:    404,
		}
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, &SessionError{
			Code:    "SESSION_EXPIRED",
			Message: "The authentication session has expired. Please start a new session.",
			HTTP:    410,
		}
	}

	if session.Status == "completed" {
		return nil, &SessionError{
			Code:    "SESSION_ALREADY_COMPLETED",
			Message: "This session has already been completed.",
			HTTP:    409,
		}
	}

	// Cap is checked before the upstream call so the user sees the same answer
	// whether or not Sekura is reachable. The counter is only incremented on a
	// real upstream response (200 or 202); upstream errors do not consume an
	// attempt because the user has done nothing wrong.
	if session.Attempts >= s.maxAttempts {
		log.Printf("[WARN] Max attempts exceeded: session_id=%s attempts=%d", authSessionID, session.Attempts)
		return nil, &SessionError{
			Code:    "MAX_ATTEMPTS_EXCEEDED",
			Message: fmt.Sprintf("Maximum verification attempts (%d) exceeded.", s.maxAttempts),
			HTTP:    429,
		}
	}

	log.Printf("[INFO] CompleteAuth: session_id=%s attempt=%d/%d", authSessionID, session.Attempts+1, s.maxAttempts)

	pollResp, ready, err := s.sekura.PollDeviceMatch(session.PollingURI)
	if err != nil {
		log.Printf("[ERROR] Polling failed: session_id=%s error=%v", authSessionID, err)
		return nil, &SessionError{
			Code:    "UPSTREAM_ERROR",
			Message: "Failed to retrieve device verification result from upstream.",
			HTTP:    502,
		}
	}

	session.Attempts++
	s.store.Update(session)

	if !ready {
		remaining := s.maxAttempts - session.Attempts
		log.Printf("[INFO] Device not yet verified: session_id=%s attempts_remaining=%d", authSessionID, remaining)
		return &AuthResult{
			IsPending: true,
			Pending: &model.PendingResponse{
				AuthSessionID:     authSessionID,
				Status:            "pending",
				Message:           "Device verification not yet complete. Ensure session_uri has been loaded on device over mobile data, then retry.",
				AttemptsRemaining: remaining,
			},
		}, nil
	}

	// Build the ALLOW/DENY decision.
	//   SimSwapResult == true  → no recent swap (safe)
	//   SimSwapResult == false → recent swap detected (deny)
	now := time.Now().UTC()
	deviceMatch := pollResp.DeviceMatch
	simSwapSafe := session.SimSwapResult

	var decision, reasonCode, reasonMessage string
	switch {
	case deviceMatch && simSwapSafe:
		decision = "ALLOW"
		reasonCode = "SIM_MATCH_SUCCESS"
		reasonMessage = "SIM verified successfully. No recent SIM swap detected."
	case deviceMatch && !simSwapSafe:
		decision = "DENY"
		reasonCode = "RECENT_SIM_SWAP"
		reasonMessage = "SIM verified but a recent SIM swap was detected. Access denied for security."
	default:
		decision = "DENY"
		reasonCode = "SIM_MISMATCH"
		reasonMessage = "Device SIM does not match registered MSISDN."
	}

	session.Status = "completed"
	session.Decision = decision
	session.ReasonCode = reasonCode
	session.ReasonMessage = reasonMessage
	session.CompletedAt = &now
	s.store.Update(session)

	log.Printf("[INFO] Auth decision: session_id=%s decision=%s reason=%s", authSessionID, decision, reasonCode)

	return &AuthResult{
		IsPending: false,
		Complete: &model.CompleteResponse{
			AuthSessionID: authSessionID,
			Decision:      decision,
			ReasonCode:    reasonCode,
			ReasonMessage: reasonMessage,
			DeviceMatch:   deviceMatch,
			SimSwapSafe:   simSwapSafe,
			CompletedAt:   now,
		},
	}, nil
}

// GetUpstreamSessionURI retrieves the upstream SessionURI for a given session ID.
// This is a non-consuming peek; use ConsumeSessionURI when actually serving the
// browser redirect so reuse is detectable.
func (s *AuthService) GetUpstreamSessionURI(authSessionID string) (string, error) {
	session, ok := s.store.Get(authSessionID)
	if !ok {
		return "", fmt.Errorf("session not found")
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		return "", fmt.Errorf("session expired")
	}
	return session.SessionURI, nil
}

// ConsumeSessionURI returns the upstream SessionURI and marks the session as
// having been redirected to. The second return is the previous redirected_at
// timestamp: nil on the first consume, non-nil on every subsequent call.
// Callers should treat a non-nil prior timestamp as "session already used".
//
// Note: the in-memory store's read-modify-write is not atomic across this
// function; under heavy concurrency two requests can both observe nil and
// both proceed. That is acceptable here because the upstream itself enforces
// single-use, so the second request will get INVALID_REQUEST from Sekura
// regardless. The marker is a UX safeguard, not a hard lock.
func (s *AuthService) ConsumeSessionURI(authSessionID string) (string, *time.Time, error) {
	session, ok := s.store.Get(authSessionID)
	if !ok {
		return "", nil, fmt.Errorf("session not found")
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		return "", nil, fmt.Errorf("session expired")
	}
	prior := session.RedirectedAt
	if prior == nil {
		now := time.Now().UTC()
		session.RedirectedAt = &now
		s.store.Update(session)
	}
	return session.SessionURI, prior, nil
}

// PollBySessionID proxies polling for the wrapped polling endpoint.
func (s *AuthService) PollBySessionID(authSessionID string) (*model.SekuraPollingResponse, bool, error) {
	session, ok := s.store.Get(authSessionID)
	if !ok {
		return nil, false, fmt.Errorf("session not found")
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, false, fmt.Errorf("session expired")
	}
	return s.sekura.PollDeviceMatch(session.PollingURI)
}
