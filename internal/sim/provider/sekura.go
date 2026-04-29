package provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"device_only/internal/sim/model"
)

// tokenSkew is subtracted from the upstream expires_in so we refresh slightly
// before Sekura would actually reject the token.
const tokenSkew = 60 * time.Second

// SekuraProvider handles all communication with the Sekura (XConnect) APIs.
type SekuraProvider struct {
	baseURL       string
	clientKey     string
	clientSecret  string
	refreshToken  string
	pollingKey    string
	pollingSecret string
	scopes        []string
	httpClient    *http.Client
	maxRetries    int
	retryDelay    time.Duration

	tokenMu        sync.Mutex
	cachedToken    string
	tokenExpiresAt time.Time
}

// NewSekuraProvider constructs a SekuraProvider with a 30-second HTTP timeout
// and the supplied retry policy. maxRetries is the number of additional
// attempts on top of the initial call (so 0 means "no retry"). retryDelayMs
// is the fixed sleep between attempts; 5xx and network errors are retried,
// 4xx and 2xx are not.
//
// pollingKey/pollingSecret are a separate Basic-auth credential pair that
// Sekura issues specifically for the polling_uri endpoint - distinct from
// clientKey/clientSecret which authenticate the token endpoint.
func NewSekuraProvider(baseURL, clientKey, clientSecret, refreshToken, pollingKey, pollingSecret string, scopes []string, maxRetries, retryDelayMs int) *SekuraProvider {
	if maxRetries < 0 {
		maxRetries = 0
	}
	if retryDelayMs < 0 {
		retryDelayMs = 0
	}
	return &SekuraProvider{
		baseURL:       baseURL,
		clientKey:     clientKey,
		clientSecret:  clientSecret,
		refreshToken:  refreshToken,
		pollingKey:    pollingKey,
		pollingSecret: pollingSecret,
		scopes:        scopes,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		maxRetries:    maxRetries,
		retryDelay:    time.Duration(retryDelayMs) * time.Millisecond,
	}
}

// basicAuth returns the Basic auth header value for client_key:client_secret.
func (p *SekuraProvider) basicAuth() string {
	raw := p.clientKey + ":" + p.clientSecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

// pollingBasicAuth returns the Basic auth header value for the polling
// endpoint, using the polling-specific credential pair.
func (p *SekuraProvider) pollingBasicAuth() string {
	raw := p.pollingKey + ":" + p.pollingSecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

// doWithRetry rebuilds the request on every attempt (so the body reader is
// always fresh) and retries on connection errors and 5xx responses up to
// maxRetries+1 total attempts. The returned response, if any, has not yet
// been read; the caller owns its Body.
func (p *SekuraProvider) doWithRetry(method, url string, body []byte, headers map[string]string) (*http.Response, error) {
	var lastErr error
	attempts := p.maxRetries + 1
	for i := 0; i < attempts; i++ {
		if i > 0 && p.retryDelay > 0 {
			time.Sleep(p.retryDelay)
		}

		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}
		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := p.httpClient.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("[WARN] Sekura attempt %d/%d failed: %v", i+1, attempts, err)
			continue
		}
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("upstream HTTP %d: %s", resp.StatusCode, string(b))
			log.Printf("[WARN] Sekura attempt %d/%d returned %d", i+1, attempts, resp.StatusCode)
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

// GetToken returns a valid Sekura access token, reusing the cached value
// while it is still within its expires_in window (minus tokenSkew). When the
// cache is cold or stale it calls /v1/token with retry-on-5xx.
func (p *SekuraProvider) GetToken() (string, error) {
	p.tokenMu.Lock()
	defer p.tokenMu.Unlock()

	if p.cachedToken != "" && time.Now().Before(p.tokenExpiresAt) {
		return p.cachedToken, nil
	}

	url := p.baseURL + "/v1/token"
	log.Printf("[INFO] Sekura API call: POST %s", url)
	start := time.Now()

	payload := map[string]any{
		"grant_type":    "refresh_token",
		"refresh_token": p.refreshToken,
	}
	if len(p.scopes) > 0 {
		payload["scopes"] = p.scopes
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	resp, err := p.doWithRetry(http.MethodPost, url, body, map[string]string{
		"Authorization": p.basicAuth(),
		"Content-Type":  "application/json",
	})
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
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("sekura token response missing access_token")
	}

	p.cachedToken = tokenResp.AccessToken
	if tokenResp.ExpiresIn > 0 {
		ttl := time.Duration(tokenResp.ExpiresIn)*time.Second - tokenSkew
		if ttl < 0 {
			ttl = 0
		}
		p.tokenExpiresAt = time.Now().Add(ttl)
	} else {
		// Conservative fallback when the upstream omits expires_in.
		p.tokenExpiresAt = time.Now().Add(5 * time.Minute)
	}

	return p.cachedToken, nil
}

// invalidateToken clears the cached token so the next GetToken call refetches.
// Useful when an insights call returns 401 indicating the token was rejected.
func (p *SekuraProvider) invalidateToken() {
	p.tokenMu.Lock()
	defer p.tokenMu.Unlock()
	p.cachedToken = ""
	p.tokenExpiresAt = time.Time{}
}

// GetInsights calls the Sekura insights endpoint for the given MSISDN. It returns
// device_match URIs, sim_swap_check data, and operator information in one call.
// On 401 the cached token is invalidated so the caller can retry with a fresh one.
func (p *SekuraProvider) GetInsights(msisdn, accessToken string) (*model.SekuraInsightsResponse, error) {
	url := p.baseURL + "/v1/insights/" + msisdn
	log.Printf("[INFO] Sekura API call: POST %s", url)
	start := time.Now()

	resp, err := p.doWithRetry(http.MethodPost, url, nil, map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Content-Type":  "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("execute insights request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[INFO] Sekura insights response: status=%d duration=%s", resp.StatusCode, time.Since(start).Round(time.Millisecond))

	if resp.StatusCode == http.StatusUnauthorized {
		p.invalidateToken()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sekura insights API returned 401: %s", string(b))
	}
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
//   - (response, true, nil)  - device has completed SIM verification
//   - (nil, false, nil)      - device has not yet loaded the session_uri (HTTP 202)
//   - (nil, false, err)      - upstream error
func (p *SekuraProvider) PollDeviceMatch(pollingURI string) (*model.SekuraPollingResponse, bool, error) {
	log.Printf("[INFO] Sekura API call: POST %s", pollingURI)
	start := time.Now()

	resp, err := p.doWithRetry(http.MethodPost, pollingURI, []byte("{}"), map[string]string{
		"Authorization": p.pollingBasicAuth(),
		"Content-Type":  "application/json",
	})
	if err != nil {
		return nil, false, fmt.Errorf("execute polling request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[INFO] Sekura polling response: status=%d duration=%s", resp.StatusCode, time.Since(start).Round(time.Millisecond))

	// 202 Accepted - device has not loaded the session_uri yet.
	if resp.StatusCode == http.StatusAccepted {
		return nil, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("[ERROR] Sekura polling non-2xx: status=%d uri=%s body=%s", resp.StatusCode, pollingURI, string(b))
		return nil, false, fmt.Errorf("sekura polling API returned HTTP %d: %s", resp.StatusCode, string(b))
	}

	var pollResp model.SekuraPollingResponse
	if err := json.NewDecoder(resp.Body).Decode(&pollResp); err != nil {
		return nil, false, fmt.Errorf("decode polling response: %w", err)
	}

	return &pollResp, true, nil
}
