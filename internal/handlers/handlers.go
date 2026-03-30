package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode"

	"device_only/internal/config"
	"device_only/internal/models"
	"device_only/internal/orchestration"
	"device_only/internal/service"
	simmodel "device_only/internal/sim/model"
	simservice "device_only/internal/sim/service"

	"github.com/google/uuid"
)

type API struct {
	deviceService   *service.DeviceService
	authService     *service.AuthService
	simAuthService  SimAuthOrchestrator
	cfg             *config.Config
	orchestrationDB *orchestration.Store
}

type SimAuthOrchestrator interface {
	StartAuth(req *simmodel.StartRequest) (*simmodel.StartResponse, error)
	CompleteAuth(authSessionID string) (*simservice.AuthResult, *simservice.SessionError)
}

func NewAPI(ds *service.DeviceService, as *service.AuthService, simAS SimAuthOrchestrator, cfg *config.Config, orchestrationDB *orchestration.Store) *API {
	return &API{
		deviceService:   ds,
		authService:     as,
		simAuthService:  simAS,
		cfg:             cfg,
		orchestrationDB: orchestrationDB,
	}
}

func (api *API) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/device/check", api.handleDeviceCheck)
	mux.HandleFunc("POST /v1/device/register", api.handleDeviceRegister)
	mux.HandleFunc("POST /v1/device/revoke", api.handleDeviceRevoke)
	mux.HandleFunc("PUT /v1/device/update", api.handleDeviceUpdate)

	mux.HandleFunc("POST /v1/auth/start", api.handleAuthStart)
	mux.HandleFunc("POST /v1/auth/complete", api.handleAuthComplete)
	mux.HandleFunc("POST /v1/auth/verify", api.handleAuthVerify)
	mux.HandleFunc("POST /v1/hybrid/start", api.handleHybridStart)
	mux.HandleFunc("POST /v1/hybrid/complete", api.handleHybridComplete)
}

func (api *API) handleDeviceRegister(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	var req models.RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid_request", err.Error(), requestID)
		return
	}

	resp, err := api.deviceService.RegisterDevice(&req)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "internal_error", err.Error(), requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, http.StatusOK, resp)
}

func (api *API) handleDeviceRevoke(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	var req models.RevokeDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid_request", err.Error(), requestID)
		return
	}

	resp, err := api.deviceService.RevokeDevice(&req)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "internal_error", err.Error(), requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, http.StatusOK, resp)
}

func (api *API) handleDeviceCheck(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	clientID := r.URL.Query().Get("client_id")
	userRef := r.URL.Query().Get("user_ref")
	deviceID := r.URL.Query().Get("device_id")

	if clientID == "" || userRef == "" {
		api.writeError(w, http.StatusBadRequest, "validation_error", "missing client_id or user_ref query parameter", requestID)
		return
	}

	resp, err := api.deviceService.CheckDeviceBinding(clientID, userRef, deviceID)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "internal_error", err.Error(), requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, http.StatusOK, resp)
}

func (api *API) handleDeviceUpdate(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	var req models.UpdateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid_request", err.Error(), requestID)
		return
	}

	resp, err := api.deviceService.UpdateDevice(&req)
	if err != nil {
		api.writeError(w, http.StatusBadRequest, "internal_error", err.Error(), requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, http.StatusOK, resp)
}

func (api *API) handleAuthStart(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	var req models.StartAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid_request", err.Error(), requestID)
		return
	}

	resp, status, err := api.processAuthStart(&req)
	if err != nil {
		api.writeError(w, status, errorCodeFor(status, err), err.Error(), requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, status, resp)
}

func (api *API) handleHybridStart(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	var req models.StartAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid_request", err.Error(), requestID)
		return
	}
	req.Mode = "hybrid"

	resp, status, err := api.processAuthStart(&req)
	if err != nil {
		api.writeError(w, status, errorCodeFor(status, err), err.Error(), requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, status, resp)
}

func (api *API) handleAuthComplete(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	var req models.CompleteAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid_request", err.Error(), requestID)
		return
	}

	resp, status, err := api.processAuthComplete(&req)
	if err != nil {
		api.writeError(w, status, errorCodeFor(status, err), err.Error(), requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, status, resp)
}

func (api *API) handleHybridComplete(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	var req models.CompleteAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid_request", err.Error(), requestID)
		return
	}

	resp, status, err := api.processAuthComplete(&req)
	if err != nil {
		api.writeError(w, status, errorCodeFor(status, err), err.Error(), requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, status, resp)
}

func (api *API) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	var req models.VerifyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid_request", err.Error(), requestID)
		return
	}

	resp, err := api.authService.VerifyToken(&req)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "internal_error", err.Error(), requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, http.StatusOK, resp)
}

func (api *API) writeError(w http.ResponseWriter, status int, errCode, message, requestID string) {
	writeJSON(w, status, models.ErrorResponse{
		Error:     errCode,
		Message:   message,
		RequestID: requestID,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (api *API) processAuthStart(req *models.StartAuthRequest) (*models.StartAuthResponse, int, error) {
	if strings.TrimSpace(req.ClientID) == "" || strings.TrimSpace(req.UserRef) == "" {
		return nil, http.StatusBadRequest, errValidation("client_id and user_ref are required")
	}

	mode := normalizeAuthMode(req.Mode)
	if req.Mode != "" && mode == "" {
		return nil, http.StatusBadRequest, errValidation("invalid mode, expected one of: device|sim|hybrid")
	}
	if mode == "" {
		mode = "device"
	}
	req.Mode = mode

	switch mode {
	case "device":
		resp, err := api.authService.StartAuth(req)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		resp.Mode = "device"
		return resp, http.StatusOK, nil
	case "sim":
		if !validMSISDN(req.MSISDN) {
			return nil, http.StatusBadRequest, errValidation("msisdn must be numeric and at least 10 digits for sim mode")
		}
		resp, err := api.startSIMOnlyAuth(req)
		if err != nil {
			return nil, http.StatusBadGateway, err
		}
		return resp, http.StatusOK, nil
	case "hybrid":
		if strings.TrimSpace(req.DeviceBindingID) == "" {
			return nil, http.StatusBadRequest, errValidation("device_binding_id is required for hybrid mode")
		}
		if strings.TrimSpace(req.DeviceInfo.DeviceID) == "" {
			return nil, http.StatusBadRequest, errValidation("device_info.device_id is required for hybrid mode")
		}
		if !validMSISDN(req.MSISDN) {
			return nil, http.StatusBadRequest, errValidation("msisdn must be numeric and at least 10 digits for hybrid mode")
		}
		resp, err := api.startHybridAuth(req)
		if err != nil {
			return nil, http.StatusBadGateway, err
		}
		return resp, http.StatusOK, nil
	default:
		return nil, http.StatusBadRequest, errValidation("unsupported mode")
	}
}

func (api *API) processAuthComplete(req *models.CompleteAuthRequest) (*models.CompleteAuthResponse, int, error) {
	if strings.TrimSpace(req.AuthSessionID) == "" {
		return nil, http.StatusBadRequest, errValidation("auth_session_id is required")
	}

	requestedMode := normalizeAuthMode(req.Mode)
	if strings.TrimSpace(req.Mode) != "" && requestedMode == "" {
		return nil, http.StatusBadRequest, errValidation("invalid mode, expected one of: device|sim|hybrid")
	}

	orchSession, ok := api.orchestrationDB.Get(req.AuthSessionID)
	if !ok {
		if requestedMode != "" && requestedMode != "device" {
			return nil, http.StatusBadRequest, errValidation("mode does not match auth session (expected device)")
		}

		// Backwards-compatible path for device-only sessions created by /v1/auth/start.
		resp, err := api.authService.CompleteAuth(req)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		resp.Mode = "device"
		return resp, http.StatusOK, nil
	}

	if time.Now().After(orchSession.ExpiresAt) {
		orchSession.Status = "FAILED"
		now := time.Now().UTC()
		orchSession.CompletedAt = &now
		api.orchestrationDB.Update(orchSession)
		return &models.CompleteAuthResponse{
			Timestamp:     time.Now().Format(time.RFC3339),
			ClientID:      orchSession.ClientID,
			Mode:          orchSession.Mode,
			AuthSessionID: orchSession.AuthSessionID,
			Decision:      "DENY",
			ReasonCode:    "SESSION_EXPIRED",
			Status:        "FAILED",
		}, http.StatusOK, nil
	}

	if requestedMode != "" && requestedMode != orchSession.Mode {
		return nil, http.StatusBadRequest, errValidation("mode does not match auth session")
	}

	switch orchSession.Mode {
	case "sim":
		return api.completeSIMOnlyAuth(req, orchSession)
	case "hybrid":
		return api.completeHybridAuth(req, orchSession)
	default:
		return nil, http.StatusInternalServerError, errValidation("orchestration session has unsupported mode")
	}
}

func (api *API) startSIMOnlyAuth(req *models.StartAuthRequest) (*models.StartAuthResponse, error) {
	simResp, err := api.simAuthService.StartAuth(&simmodel.StartRequest{
		ClientID: req.ClientID,
		UserRef:  req.UserRef,
		MSISDN:   req.MSISDN,
		Action:   req.Action,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	authSessionID := "auth_" + uuid.New().String()
	api.orchestrationDB.Set(&orchestration.Session{
		AuthSessionID:    authSessionID,
		ClientID:         req.ClientID,
		UserRef:          req.UserRef,
		Mode:             "sim",
		SimAuthSessionID: simResp.AuthSessionID,
		Status:           "PENDING",
		CreatedAt:        now,
		ExpiresAt:        now.Add(time.Duration(simResp.ExpiresIn) * time.Second),
	})

	return &models.StartAuthResponse{
		Timestamp:     time.Now().Format(time.RFC3339),
		ClientID:      req.ClientID,
		Mode:          "sim",
		AuthSessionID: authSessionID,
		NextStep:      "SIM_CHALLENGE_REQUIRED",
		Sim: &models.SimChallenge{
			AuthSessionID:    simResp.AuthSessionID,
			SessionURI:       simResp.SessionURI,
			ExpiresInSeconds: simResp.ExpiresIn,
			Instructions:     simResp.Instructions,
		},
		Status: "PENDING",
	}, nil
}

func (api *API) startHybridAuth(req *models.StartAuthRequest) (*models.StartAuthResponse, error) {
	deviceResp, err := api.authService.StartAuth(req)
	if err != nil {
		return nil, err
	}
	if deviceResp.Status != "PENDING" || deviceResp.Device == nil || deviceResp.AuthSessionID == "" {
		deviceResp.Mode = "hybrid"
		return deviceResp, nil
	}

	simResp, err := api.simAuthService.StartAuth(&simmodel.StartRequest{
		ClientID: req.ClientID,
		UserRef:  req.UserRef,
		MSISDN:   req.MSISDN,
		Action:   req.Action,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	ttlSeconds := api.cfg.ChallengeExpirySeconds
	if simResp.ExpiresIn < ttlSeconds {
		ttlSeconds = simResp.ExpiresIn
	}
	authSessionID := "auth_" + uuid.New().String()

	api.orchestrationDB.Set(&orchestration.Session{
		AuthSessionID:       authSessionID,
		ClientID:            req.ClientID,
		UserRef:             req.UserRef,
		Mode:                "hybrid",
		DeviceAuthSessionID: deviceResp.AuthSessionID,
		DeviceChallengeID:   deviceResp.Device.ChallengeID,
		SimAuthSessionID:    simResp.AuthSessionID,
		Status:              "PENDING",
		CreatedAt:           now,
		ExpiresAt:           now.Add(time.Duration(ttlSeconds) * time.Second),
	})

	return &models.StartAuthResponse{
		Timestamp:     time.Now().Format(time.RFC3339),
		ClientID:      req.ClientID,
		Mode:          "hybrid",
		AuthSessionID: authSessionID,
		NextStep:      "SIM_AND_DEVICE_REQUIRED",
		Device:        deviceResp.Device,
		Sim: &models.SimChallenge{
			AuthSessionID:    simResp.AuthSessionID,
			SessionURI:       simResp.SessionURI,
			ExpiresInSeconds: simResp.ExpiresIn,
			Instructions:     simResp.Instructions,
		},
		Status: "PENDING",
	}, nil
}

func (api *API) completeSIMOnlyAuth(req *models.CompleteAuthRequest, orchSession *orchestration.Session) (*models.CompleteAuthResponse, int, error) {
	result, sessionErr := api.simAuthService.CompleteAuth(orchSession.SimAuthSessionID)
	if sessionErr != nil {
		return &models.CompleteAuthResponse{
			Timestamp:     time.Now().Format(time.RFC3339),
			ClientID:      orchSession.ClientID,
			Mode:          "sim",
			AuthSessionID: orchSession.AuthSessionID,
			Decision:      "DENY",
			ReasonCode:    sessionErr.Code,
			ReasonMessage: sessionErr.Message,
			Status:        "FAILED",
		}, sessionErr.HTTP, nil
	}

	if result.IsPending {
		return &models.CompleteAuthResponse{
			Timestamp:         time.Now().Format(time.RFC3339),
			ClientID:          orchSession.ClientID,
			Mode:              "sim",
			AuthSessionID:     orchSession.AuthSessionID,
			Decision:          "PENDING",
			ReasonCode:        "SIM_PENDING",
			ReasonMessage:     result.Pending.Message,
			NextStep:          "SIM_CHALLENGE_REQUIRED",
			AttemptsRemaining: result.Pending.AttemptsRemaining,
			Sim: &models.SimDecision{
				Decision:      "PENDING",
				ReasonCode:    "SIM_PENDING",
				ReasonMessage: result.Pending.Message,
				Status:        "PENDING",
			},
			Status: "PENDING",
		}, http.StatusAccepted, nil
	}

	if result.Complete.Decision != "ALLOW" {
		now := time.Now().UTC()
		orchSession.Status = "FAILED"
		orchSession.CompletedAt = &now
		api.orchestrationDB.Update(orchSession)
		return &models.CompleteAuthResponse{
			Timestamp:     time.Now().Format(time.RFC3339),
			ClientID:      orchSession.ClientID,
			Mode:          "sim",
			AuthSessionID: orchSession.AuthSessionID,
			Decision:      "DENY",
			ReasonCode:    result.Complete.ReasonCode,
			ReasonMessage: result.Complete.ReasonMessage,
			Sim: &models.SimDecision{
				Decision:      result.Complete.Decision,
				ReasonCode:    result.Complete.ReasonCode,
				ReasonMessage: result.Complete.ReasonMessage,
				Status:        "COMPLETED",
			},
			Status: "FAILED",
		}, http.StatusOK, nil
	}

	tokenID, err := api.authService.CreateContextToken(orchSession.UserRef)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	now := time.Now().UTC()
	orchSession.Status = "SUCCESS"
	orchSession.CompletedAt = &now
	api.orchestrationDB.Update(orchSession)

	return &models.CompleteAuthResponse{
		Timestamp:        time.Now().Format(time.RFC3339),
		ClientID:         orchSession.ClientID,
		Mode:             "sim",
		AuthSessionID:    orchSession.AuthSessionID,
		Decision:         "ALLOW",
		ReasonCode:       result.Complete.ReasonCode,
		ReasonMessage:    result.Complete.ReasonMessage,
		AuthContextToken: tokenID,
		ExpiresInSeconds: api.cfg.AuthTokenExpirySeconds,
		Sim: &models.SimDecision{
			Decision:      result.Complete.Decision,
			ReasonCode:    result.Complete.ReasonCode,
			ReasonMessage: result.Complete.ReasonMessage,
			Status:        "COMPLETED",
		},
		Status: "SUCCESS",
	}, http.StatusOK, nil
}

func (api *API) completeHybridAuth(req *models.CompleteAuthRequest, orchSession *orchestration.Session) (*models.CompleteAuthResponse, int, error) {
	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		clientID = orchSession.ClientID
	}

	result, sessionErr := api.simAuthService.CompleteAuth(orchSession.SimAuthSessionID)
	if sessionErr != nil {
		return &models.CompleteAuthResponse{
			Timestamp:     time.Now().Format(time.RFC3339),
			ClientID:      clientID,
			Mode:          "hybrid",
			AuthSessionID: orchSession.AuthSessionID,
			Decision:      "DENY",
			ReasonCode:    sessionErr.Code,
			ReasonMessage: sessionErr.Message,
			Status:        "FAILED",
		}, sessionErr.HTTP, nil
	}

	if result.IsPending {
		return &models.CompleteAuthResponse{
			Timestamp:         time.Now().Format(time.RFC3339),
			ClientID:          clientID,
			Mode:              "hybrid",
			AuthSessionID:     orchSession.AuthSessionID,
			Decision:          "PENDING",
			ReasonCode:        "SIM_PENDING",
			ReasonMessage:     result.Pending.Message,
			NextStep:          "SIM_CHALLENGE_REQUIRED",
			AttemptsRemaining: result.Pending.AttemptsRemaining,
			Sim: &models.SimDecision{
				Decision:      "PENDING",
				ReasonCode:    "SIM_PENDING",
				ReasonMessage: result.Pending.Message,
				Status:        "PENDING",
			},
			Status: "PENDING",
		}, http.StatusAccepted, nil
	}

	if result.Complete.Decision != "ALLOW" {
		now := time.Now().UTC()
		orchSession.Status = "FAILED"
		orchSession.CompletedAt = &now
		api.orchestrationDB.Update(orchSession)
		return &models.CompleteAuthResponse{
			Timestamp:     time.Now().Format(time.RFC3339),
			ClientID:      clientID,
			Mode:          "hybrid",
			AuthSessionID: orchSession.AuthSessionID,
			Decision:      "DENY",
			ReasonCode:    result.Complete.ReasonCode,
			ReasonMessage: result.Complete.ReasonMessage,
			Sim: &models.SimDecision{
				Decision:      result.Complete.Decision,
				ReasonCode:    result.Complete.ReasonCode,
				ReasonMessage: result.Complete.ReasonMessage,
				Status:        "COMPLETED",
			},
			Status: "FAILED",
		}, http.StatusOK, nil
	}

	if strings.TrimSpace(req.DeviceSignature) == "" {
		return nil, http.StatusBadRequest, errValidation("device_signature is required for hybrid completion")
	}

	deviceResp, err := api.authService.CompleteAuth(&models.CompleteAuthRequest{
		ClientID:        clientID,
		AuthSessionID:   orchSession.DeviceAuthSessionID,
		ChallengeID:     orchSession.DeviceChallengeID,
		DeviceSignature: req.DeviceSignature,
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	deviceResp.Mode = "hybrid"
	deviceResp.AuthSessionID = orchSession.AuthSessionID
	deviceResp.Sim = &models.SimDecision{
		Decision:      result.Complete.Decision,
		ReasonCode:    result.Complete.ReasonCode,
		ReasonMessage: result.Complete.ReasonMessage,
		Status:        "COMPLETED",
	}

	now := time.Now().UTC()
	if deviceResp.Decision == "ALLOW" {
		orchSession.Status = "SUCCESS"
		orchSession.CompletedAt = &now
		api.orchestrationDB.Update(orchSession)
		deviceResp.ReasonCode = "HYBRID_SIM_AND_DEVICE_VALID"
		return deviceResp, http.StatusOK, nil
	}

	orchSession.Status = "FAILED"
	orchSession.CompletedAt = &now
	api.orchestrationDB.Update(orchSession)
	return deviceResp, http.StatusOK, nil
}

func validMSISDN(s string) bool {
	s = strings.TrimSpace(s)
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

func normalizeAuthMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "device":
		return "device"
	case "sim":
		return "sim"
	case "hybrid":
		return "hybrid"
	default:
		return ""
	}
}

func errValidation(msg string) error {
	return &validationError{msg: msg}
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string {
	return e.msg
}

func errorCodeFor(status int, err error) string {
	var vErr *validationError
	if errors.As(err, &vErr) {
		return "validation_error"
	}
	if status == http.StatusBadGateway {
		return "upstream_error"
	}
	if status >= 400 && status < 500 {
		return "invalid_request"
	}
	return "internal_error"
}
