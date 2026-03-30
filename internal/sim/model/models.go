package model

import "time"

// Session holds all state for an active authentication session.
type Session struct {
	AuthSessionID string
	ClientID      string
	UserRef       string
	MSISDN        string
	Action        string
	SessionURI    string
	PollingURI    string
	SimSwapResult bool
	SimSwapDate   string
	Status        string // pending | completed | expired
	Decision      string // ALLOW | DENY | ""
	ReasonCode    string
	ReasonMessage string
	Attempts      int
	CreatedAt     time.Time
	ExpiresAt     time.Time
	CompletedAt   *time.Time
}

// StartRequest is the request body for POST /v1/auth/start.
type StartRequest struct {
	ClientID string `json:"client_id"`
	UserRef  string `json:"user_ref"`
	MSISDN   string `json:"msisdn"`
	Action   string `json:"action"`
}

// StartResponse is the success response for POST /v1/auth/start.
type StartResponse struct {
	RequestID      string                `json:"request_id"`
	AuthSessionID  string                `json:"auth_session_id"`
	ExpiresIn      int                   `json:"expires_in"`
	NextStep       string                `json:"next_step"`
	SessionURI     string                `json:"session_uri"`
	Instructions   string                `json:"instructions"`
	SimSwapCheck   *SimSwapCheckResult   `json:"sim_swap_check,omitempty"`
	OperatorLookup *OperatorLookupResult `json:"operator_lookup,omitempty"`
}

// CompleteRequest is the request body for POST /v1/auth/complete.
type CompleteRequest struct {
	AuthSessionID string `json:"auth_session_id"`
}

// CompleteResponse is the final decision response for POST /v1/auth/complete.
type CompleteResponse struct {
	RequestID     string    `json:"request_id"`
	AuthSessionID string    `json:"auth_session_id"`
	Decision      string    `json:"decision"`
	ReasonCode    string    `json:"reason_code"`
	ReasonMessage string    `json:"reason_message"`
	DeviceMatch   bool      `json:"device_match"`
	SimSwapSafe   bool      `json:"sim_swap_safe"`
	CompletedAt   time.Time `json:"completed_at"`
}

// PendingResponse is returned when the device has not yet loaded the session_uri.
type PendingResponse struct {
	RequestID         string `json:"request_id"`
	AuthSessionID     string `json:"auth_session_id"`
	Status            string `json:"status"`
	Message           string `json:"message"`
	AttemptsRemaining int    `json:"attempts_remaining"`
}

// ErrorResponse is the standard error envelope for all error responses.
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// SekuraTokenResponse is the response from the Sekura token endpoint.
type SekuraTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// SekuraInsightsResponse is the combined response from the Sekura insights endpoint.
type SekuraInsightsResponse struct {
	DeviceMatch struct {
		SessionURI string `json:"session_uri"`
		PollingURI string `json:"polling_uri"`
	} `json:"device_match"`
	SimSwapCheck   SimSwapCheckResult   `json:"sim_swap_check"`
	OperatorLookup OperatorLookupResult `json:"operator_lookup"`
}

type SimSwapCheckResult struct {
	Result  bool   `json:"result"`
	Date    string `json:"date"`
	Seconds int    `json:"seconds"`
}

type OperatorLookupResult struct {
	RegionCode   string `json:"regionCode"`
	OperatorName string `json:"operatorName"`
	MCC          string `json:"mcc"`
	MNC          string `json:"mnc"`
}

// SekuraPollingResponse is the response from the Sekura device-match polling endpoint.
type SekuraPollingResponse struct {
	MSISDN      string `json:"msisdn"`
	DeviceMatch bool   `json:"device_match"`
	RemoteAddr  string `json:"remote_addr"`
	UserAgent   string `json:"user_agent"`
}
