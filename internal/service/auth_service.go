package service

import (
	"device_only/internal/config"
	"device_only/internal/crypto"
	"device_only/internal/models"
	"device_only/internal/repository"
	"errors"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	repo repository.Repository
	cfg  *config.Config
}

func NewAuthService(repo repository.Repository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) StartAuth(req *models.StartAuthRequest) (*models.StartAuthResponse, error) {
	device, err := s.repo.GetDevice(req.DeviceBindingID)
	if err != nil || device.Status != "ACTIVE" || device.ClientID != req.ClientID || device.UserRef != req.UserRef || device.DeviceID != req.DeviceInfo.DeviceID {
		return &models.StartAuthResponse{
			RequestID: "req_" + uuid.New().String(),
			Timestamp: time.Now().Format(time.RFC3339),
			ClientID:  req.ClientID,
			Status:    "DEVICE_NOT_MATCH",
		}, nil
	}

	challenge, err := crypto.GenerateChallenge()
	if err != nil {
		return nil, err
	}

	sessionID := "auth_" + uuid.New().String()
	challengeID := "chal_" + uuid.New().String()

	expiresAt := time.Now().Add(time.Duration(s.cfg.ChallengeExpirySeconds) * time.Second)

	session := &models.AuthSession{
		AuthSessionID:   sessionID,
		UserRef:         req.UserRef,
		Challenge:       challenge,
		ChallengeID:     challengeID,
		DeviceBindingID: req.DeviceBindingID,
		ExpiresAt:       expiresAt,
		Status:          "PENDING",
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateAuthSession(session); err != nil {
		return nil, err
	}

	_ = s.repo.LogAudit(&models.AuditLog{
		UserRef:   req.UserRef,
		Action:    "AUTH_START",
		Decision:  "PENDING",
		IPAddress: req.DeviceInfo.IPAddress,
		DeviceID:  req.DeviceInfo.DeviceID,
	})

	return &models.StartAuthResponse{
		RequestID:     "req_" + uuid.New().String(),
		Timestamp:     time.Now().Format(time.RFC3339),
		ClientID:      req.ClientID,
		Mode:          "device",
		AuthSessionID: sessionID,
		NextStep:      "DEVICE_REQUIRED",
		Device: &models.DeviceChallenge{
			ChallengeID:      challengeID,
			Challenge:        challenge,
			ExpiresInSeconds: s.cfg.ChallengeExpirySeconds,
		},
		Status: "PENDING",
	}, nil
}

func (s *AuthService) CompleteAuth(req *models.CompleteAuthRequest) (*models.CompleteAuthResponse, error) {
	session, err := s.repo.GetAuthSession(req.AuthSessionID)
	if err != nil {
		return nil, errors.New("invalid session_id")
	}

	audit := &models.AuditLog{
		UserRef: session.UserRef,
		Action:  "AUTH_COMPLETE",
	}
	defer func() { _ = s.repo.LogAudit(audit) }() // Best-effort write

	if time.Now().After(session.ExpiresAt) {
		audit.Decision = "DENY"
		_ = s.repo.CompleteAuthSession(session.AuthSessionID, "EXPIRED")
		return s.denyResponse(req.ClientID, session.AuthSessionID, "SESSION_EXPIRED"), nil
	}

	if session.Status != "PENDING" || session.ChallengeID != req.ChallengeID {
		audit.Decision = "DENY"
		return s.denyResponse(req.ClientID, session.AuthSessionID, "INVALID_CHALLENGE"), nil
	}

	device, err := s.repo.GetDevice(session.DeviceBindingID)
	if err != nil || device.Status != "ACTIVE" {
		audit.Decision = "DENY"
		return s.denyResponse(req.ClientID, session.AuthSessionID, "DEVICE_NOT_FOUND"), nil
	}

	valid := false
	if s.cfg.DevMode {
		// Verify using simulation for DEV_MODE or allow a specific UI bypass signature
		if req.DeviceSignature == "bypass_signature" || req.DeviceSignature == "demo_bypass_signature" {
			valid = true // UI Demo bypass
		} else {
			validStr, errStr := crypto.VerifySignature(device.PublicKey, session.Challenge, req.DeviceSignature)
			if errStr != nil {
				valid = false
			} else {
				valid = validStr
			}
		}
	} else {
		// Strict production verification
		validStr, errStr := crypto.VerifySignature(device.PublicKey, session.Challenge, req.DeviceSignature)
		if errStr != nil {
			valid = false
		} else {
			valid = validStr
		}
	}

	if !valid {
		audit.Decision = "DENY"
		_ = s.repo.CompleteAuthSession(session.AuthSessionID, "FAILED")
		return s.denyResponse(req.ClientID, session.AuthSessionID, "INVALID_SIGNATURE"), nil
	}

	// Signifies one-time-use logic. Invalidate session.
	_ = s.repo.CompleteAuthSession(session.AuthSessionID, "SUCCESS")

	tokenID, err := s.CreateContextToken(session.UserRef)
	if err != nil {
		return nil, errors.New("failed generating context token")
	}

	audit.Decision = "ALLOW"

	return &models.CompleteAuthResponse{
		RequestID:        "req_" + uuid.New().String(),
		Timestamp:        time.Now().Format(time.RFC3339),
		ClientID:         req.ClientID,
		Mode:             "device",
		AuthSessionID:    session.AuthSessionID,
		Decision:         "ALLOW",
		ReasonCode:       "DEVICE_SIGNATURE_VALID",
		AuthContextToken: tokenID,
		ExpiresInSeconds: s.cfg.AuthTokenExpirySeconds,
		Status:           "SUCCESS",
	}, nil
}

func (s *AuthService) VerifyToken(req *models.VerifyTokenRequest) (*models.VerifyTokenResponse, error) {
	token, err := s.repo.GetAuthContextToken(req.AuthContextToken)
	if err != nil || token.Status != "ACTIVE" {
		return &models.VerifyTokenResponse{
			RequestID: "req_" + uuid.New().String(),
			Timestamp: time.Now().Format(time.RFC3339),
			ClientID:  req.ClientID,
			Valid:     false,
			Status:    "INVALID",
		}, nil
	}

	if time.Now().After(token.ExpiresAt) {
		return &models.VerifyTokenResponse{
			RequestID: "req_" + uuid.New().String(),
			Timestamp: time.Now().Format(time.RFC3339),
			ClientID:  req.ClientID,
			Valid:     false,
			Status:    "EXPIRED",
		}, nil
	}

	remaining := int(token.ExpiresAt.Sub(time.Now()).Seconds())

	return &models.VerifyTokenResponse{
		RequestID:        "req_" + uuid.New().String(),
		Timestamp:        time.Now().Format(time.RFC3339),
		ClientID:         req.ClientID,
		Valid:            true,
		ExpiresInSeconds: remaining,
		Status:           "ACTIVE",
	}, nil
}

func (s *AuthService) denyResponse(clientID, sessionID, reason string) *models.CompleteAuthResponse {
	return &models.CompleteAuthResponse{
		RequestID:     "req_" + uuid.New().String(),
		Timestamp:     time.Now().Format(time.RFC3339),
		ClientID:      clientID,
		Mode:          "device",
		AuthSessionID: sessionID,
		Decision:      "DENY",
		ReasonCode:    reason,
		Status:        "FAILED",
	}
}

func (s *AuthService) CreateContextToken(userRef string) (string, error) {
	tokenID := "ctx_" + uuid.New().String()
	token := &models.AuthContextToken{
		Token:     tokenID,
		UserRef:   userRef,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.AuthTokenExpirySeconds) * time.Second),
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
	if err := s.repo.CreateAuthContextToken(token); err != nil {
		return "", err
	}
	return tokenID, nil
}
