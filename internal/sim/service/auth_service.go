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
}

// NewAuthService wires dependencies and returns a ready-to-use AuthService.
func NewAuthService(
	store *store.SessionStore,
	sekura *provider.SekuraProvider,
	baseURL string,
	sessionTTLSeconds int,
	maxAttempts int,
) *AuthService {
	return &AuthService{
		store:       store,
		sekura:      sekura,
		baseURL:     strings.TrimRight(baseURL, "/"),
		sessionTTL:  time.Duration(sessionTTLSeconds) * time.Second,
		maxAttempts: maxAttempts,
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
		log.Printf("[ERROR] Sekura GetToken failed: %v. Proceeding with mock token for demo.", err)
		token = "mock_token"
	}

	insights, err := s.sekura.GetInsights(req.MSISDN, token)
	if err != nil {
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
	wrappedPollingURI := fmt.Sprintf("%s/v1/sim/poll/%s", s.baseURL, sessionID)

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

	// Increment and persist the attempt counter before calling upstream so a
	// failed/errored call still counts against the limit.
	session.Attempts++
	s.store.Update(session)

	if session.Attempts > s.maxAttempts {
		log.Printf("[WARN] Max attempts exceeded: session_id=%s attempts=%d", authSessionID, session.Attempts)
		return nil, &SessionError{
			Code:    "MAX_ATTEMPTS_EXCEEDED",
			Message: fmt.Sprintf("Maximum verification attempts (%d) exceeded.", s.maxAttempts),
			HTTP:    429,
		}
	}

	log.Printf("[INFO] CompleteAuth: session_id=%s attempt=%d/%d", authSessionID, session.Attempts, s.maxAttempts)

	pollResp, ready, err := s.sekura.PollDeviceMatch(session.PollingURI)
	if err != nil {
		log.Printf("[ERROR] Polling failed: session_id=%s error=%v", authSessionID, err)
		return nil, &SessionError{
			Code:    "UPSTREAM_ERROR",
			Message: "Failed to retrieve device verification result from upstream.",
			HTTP:    502,
		}
	}

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
