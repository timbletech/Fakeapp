package models

import "time"

// DB Models
type Device struct {
	DeviceBindingID string
	ClientID        string
	UserRef         string
	PublicKey       string
	DeviceID        string
	Platform        string
	DeviceModel     string
	OSVersion       string
	CreatedAt       time.Time
	Status          string
}

type AuthSession struct {
	AuthSessionID   string
	UserRef         string
	Challenge       string
	ChallengeID     string
	DeviceBindingID string
	ExpiresAt       time.Time
	Status          string
	CreatedAt       time.Time
}

type AuthContextToken struct {
	Token     string
	UserRef   string
	ExpiresAt time.Time
	Status    string
	CreatedAt time.Time
}

type AuditLog struct {
	ID        int
	UserRef   string
	Action    string
	Decision  string
	IPAddress string
	DeviceID  string
	CreatedAt time.Time
}

// REST JSON API Requests & Responses

// /v1/device/register
type DeviceInfo struct {
	DeviceID    string `json:"device_id"`
	Platform    string `json:"platform"`
	AppVersion  string `json:"app_version,omitempty"`
	DeviceModel string `json:"device_model,omitempty"`
	OSVersion   string `json:"os_version,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
}

type RegisterDeviceRequest struct {
	ClientID   string     `json:"client_id"`
	UserRef    string     `json:"user_ref"`
	DeviceInfo DeviceInfo `json:"device_info"`
	PublicKey  string     `json:"public_key"`
}

type RegisterDeviceResponse struct {
	RequestID       string `json:"request_id"`
	Timestamp       string `json:"timestamp"`
	ClientID        string `json:"client_id"`
	DeviceBindingID string `json:"device_binding_id"`
	Status          string `json:"status"`
}

// /v1/auth/start
type StartAuthRequest struct {
	ClientID        string     `json:"client_id"`
	UserRef         string     `json:"user_ref"`
	Action          string     `json:"action"`
	Mode            string     `json:"mode"`
	MSISDN          string     `json:"msisdn,omitempty"`
	DeviceBindingID string     `json:"device_binding_id"`
	DeviceInfo      DeviceInfo `json:"device_info"`
}

type DeviceChallenge struct {
	ChallengeID      string `json:"challenge_id"`
	Challenge        string `json:"challenge"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type StartAuthResponse struct {
	RequestID     string           `json:"request_id"`
	Timestamp     string           `json:"timestamp"`
	ClientID      string           `json:"client_id"`
	Mode          string           `json:"mode,omitempty"`
	AuthSessionID string           `json:"auth_session_id"`
	NextStep      string           `json:"next_step"`
	Device        *DeviceChallenge `json:"device,omitempty"`
	Sim           *SimChallenge    `json:"sim,omitempty"`
	Status        string           `json:"status"`
}

type SimChallenge struct {
	AuthSessionID    string `json:"auth_session_id"`
	SessionURI       string `json:"session_uri"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
	Instructions     string `json:"instructions,omitempty"`
}

// /v1/auth/complete
type CompleteAuthRequest struct {
	ClientID        string `json:"client_id"`
	Mode            string `json:"mode,omitempty"`
	AuthSessionID   string `json:"auth_session_id"`
	ChallengeID     string `json:"challenge_id"`
	DeviceSignature string `json:"device_signature"`
}

type CompleteAuthResponse struct {
	RequestID         string       `json:"request_id"`
	Timestamp         string       `json:"timestamp"`
	ClientID          string       `json:"client_id"`
	Mode              string       `json:"mode,omitempty"`
	AuthSessionID     string       `json:"auth_session_id"`
	Decision          string       `json:"decision"`
	ReasonCode        string       `json:"reason_code"`
	ReasonMessage     string       `json:"reason_message,omitempty"`
	NextStep          string       `json:"next_step,omitempty"`
	AttemptsRemaining int          `json:"attempts_remaining,omitempty"`
	Sim               *SimDecision `json:"sim,omitempty"`
	AuthContextToken  string       `json:"auth_context_token,omitempty"`
	ExpiresInSeconds  int          `json:"expires_in_seconds,omitempty"`
	Status            string       `json:"status"`
}

type SimDecision struct {
	Decision      string `json:"decision"`
	ReasonCode    string `json:"reason_code"`
	ReasonMessage string `json:"reason_message,omitempty"`
	Status        string `json:"status"`
}

// /v1/auth/verify
type VerifyTokenRequest struct {
	ClientID         string `json:"client_id"`
	AuthContextToken string `json:"auth_context_token"`
}

type VerifyTokenResponse struct {
	RequestID        string `json:"request_id"`
	Timestamp        string `json:"timestamp"`
	ClientID         string `json:"client_id"`
	Valid            bool   `json:"valid"`
	ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	Status           string `json:"status"`
}

// Optional: Revoke
type RevokeDeviceRequest struct {
	ClientID        string `json:"client_id"`
	UserRef         string `json:"user_ref"`
	DeviceBindingID string `json:"device_binding_id"`
}

type RevokeDeviceResponse struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
}

// /v1/device/check
type CheckDeviceResponse struct {
	RequestID       string `json:"request_id"`
	Timestamp       string `json:"timestamp"`
	HasActiveDevice bool   `json:"has_active_device"`
	DeviceBindingID string `json:"device_binding_id,omitempty"` // Only return the primary if exists
	Message         string `json:"message,omitempty"`
	Status          string `json:"status"`
}

// /v1/device/update
type UpdateDeviceRequest struct {
	ClientID   string     `json:"client_id"`
	UserRef    string     `json:"user_ref"`
	DeviceInfo DeviceInfo `json:"device_info"`
	PublicKey  string     `json:"public_key"`
}

type UpdateDeviceResponse struct {
	RequestID       string `json:"request_id"`
	Timestamp       string `json:"timestamp"`
	ClientID        string `json:"client_id"`
	DeviceBindingID string `json:"device_binding_id"`
	Status          string `json:"status"`
}

// ErrorResponse is the standard error envelope for all error responses.
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}
