package provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"device_only/internal/sim/model"
)

// SekuraProvider handles all communication with the Sekura (XConnect) APIs.
type SekuraProvider struct {
	baseURL      string
	clientKey    string
	clientSecret string
	refreshToken string
	httpClient   *http.Client
}

// NewSekuraProvider constructs a SekuraProvider with a 30-second HTTP timeout.
func NewSekuraProvider(baseURL, clientKey, clientSecret, refreshToken string) *SekuraProvider {
	return &SekuraProvider{
		baseURL:      baseURL,
		clientKey:    clientKey,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// basicAuth returns the Basic auth header value for client_key:client_secret.
func (p *SekuraProvider) basicAuth() string {
	raw := p.clientKey + ":" + p.clientSecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

// GetToken calls the Sekura token endpoint and returns a fresh access token.
// A new token is fetched on every request — no caching.
func (p *SekuraProvider) GetToken() (string, error) {
	url := p.baseURL + "/v1/token"
	log.Printf("[INFO] Sekura API call: POST %s", url)
	start := time.Now()

	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": p.refreshToken,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Authorization", p.basicAuth())
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[INFO] Sekura token response: status=%d duration=%s", resp.StatusCode, time.Since(start).Round(time.Millisecond))

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("sekura token API returned HTTP %d: %s", resp.StatusCode, string(b))
	}

	var tokenResp model.SekuraTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// GetInsights calls the Sekura insights endpoint for the given MSISDN. It returns
// device_match URIs, sim_swap_check data, and operator information in one call.
func (p *SekuraProvider) GetInsights(msisdn, accessToken string) (*model.SekuraInsightsResponse, error) {
	url := p.baseURL + "/v1/insights/" + msisdn
	log.Printf("[INFO] Sekura API call: POST %s", url)
	start := time.Now()

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build insights request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute insights request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[INFO] Sekura insights response: status=%d duration=%s", resp.StatusCode, time.Since(start).Round(time.Millisecond))

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sekura insights API returned HTTP %d: %s", resp.StatusCode, string(b))
	}

	var insightsResp model.SekuraInsightsResponse
	if err := json.NewDecoder(resp.Body).Decode(&insightsResp); err != nil {
		return nil, fmt.Errorf("decode insights response: %w", err)
	}

	if insightsResp.DeviceMatch.SessionURI == "" || insightsResp.DeviceMatch.PollingURI == "" {
		return nil, fmt.Errorf("sekura insights response missing device_match URIs")
	}

	return &insightsResp, nil
}

// PollDeviceMatch calls the full polling_uri (not prefixed with base URL) using
// Basic auth. It returns:
//   - (response, true, nil)  — device has completed SIM verification
//   - (nil, false, nil)      — device has not yet loaded the session_uri (HTTP 202)
//   - (nil, false, err)      — upstream error
func (p *SekuraProvider) PollDeviceMatch(pollingURI string) (*model.SekuraPollingResponse, bool, error) {
	log.Printf("[INFO] Sekura API call: GET %s", pollingURI)
	start := time.Now()

	req, err := http.NewRequest(http.MethodGet, pollingURI, nil)
	if err != nil {
		return nil, false, fmt.Errorf("build polling request: %w", err)
	}
	req.Header.Set("Authorization", p.basicAuth())
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("execute polling request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[INFO] Sekura polling response: status=%d duration=%s", resp.StatusCode, time.Since(start).Round(time.Millisecond))

	// 202 Accepted — device has not loaded the session_uri yet.
	if resp.StatusCode == http.StatusAccepted {
		return nil, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("sekura polling API returned HTTP %d: %s", resp.StatusCode, string(b))
	}

	var pollResp model.SekuraPollingResponse
	if err := json.NewDecoder(resp.Body).Decode(&pollResp); err != nil {
		return nil, false, fmt.Errorf("decode polling response: %w", err)
	}

	return &pollResp, true, nil
}
