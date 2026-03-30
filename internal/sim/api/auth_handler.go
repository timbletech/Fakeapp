package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"unicode"

	"device_only/internal/sim/model"
	"device_only/internal/sim/service"

	"github.com/google/uuid"
)

// AuthHandler exposes the two Timble authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Start handles POST /v1/auth/start.
func (h *AuthHandler) Start(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST is supported.", requestID)
		return
	}

	var req model.StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.", requestID)
		return
	}

	log.Printf("[INFO] POST /v1/sim/start client_id=%s user_ref=%s", req.ClientID, req.UserRef)

	if strings.TrimSpace(req.ClientID) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'client_id' is required.", requestID)
		return
	}
	if strings.TrimSpace(req.UserRef) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'user_ref' is required.", requestID)
		return
	}
	if strings.TrimSpace(req.MSISDN) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'msisdn' is required.", requestID)
		return
	}
	if !validMSISDN(req.MSISDN) {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'msisdn' must be numeric and at least 10 digits.", requestID)
		return
	}

	resp, err := h.authService.StartAuth(&req)
	if err != nil {
		log.Printf("[ERROR] StartAuth failed: %v", err)
		writeError(w, http.StatusBadGateway, "upstream_error", "Failed to initialize verification with upstream provider.", requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, http.StatusOK, resp)
}

// Complete handles POST /v1/auth/complete.
func (h *AuthHandler) Complete(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST is supported.", requestID)
		return
	}

	var req model.CompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.", requestID)
		return
	}

	log.Printf("[INFO] POST /v1/sim/complete auth_session_id=%s", req.AuthSessionID)

	if strings.TrimSpace(req.AuthSessionID) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'auth_session_id' is required.", requestID)
		return
	}

	result, sessErr := h.authService.CompleteAuth(req.AuthSessionID)
	if sessErr != nil {
		log.Printf("[WARN] CompleteAuth session error: code=%s session_id=%s", sessErr.Code, req.AuthSessionID)
		writeError(w, sessErr.HTTP, strings.ToLower(sessErr.Code), sessErr.Message, requestID)
		return
	}

	if result.IsPending {
		result.Pending.RequestID = requestID
		writeJSON(w, http.StatusAccepted, result.Pending)
		return
	}

	result.Complete.RequestID = requestID
	writeJSON(w, http.StatusOK, result.Complete)
}

// Redirect handles GET /v1/sim/redirect/{session_id}.
// It looks up the session and issues a 302 Found redirect to the upstream SessionURI.
func (h *AuthHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Simple path extraction assuming path is exactly /v1/sim/redirect/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid redirect URL", http.StatusBadRequest)
		return
	}
	sessionID := pathParts[4]

	sessURI, err := h.authService.GetUpstreamSessionURI(sessionID)
	if err != nil {
		http.Error(w, "Session not found or expired", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, sessURI, http.StatusFound)
}

// validMSISDN returns true when s is all digits and at least 10 characters long.
func validMSISDN(s string) bool {
	if len(s) < 10 {
		return false
	}
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, errCode, message, requestID string) {
	writeJSON(w, status, model.ErrorResponse{
		Error:     errCode,
		Message:   message,
		RequestID: requestID,
	})
}
