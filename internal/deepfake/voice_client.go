package deepfake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// VoiceClient wraps the voice analysis service HTTP API.
type VoiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewVoiceClient creates a VoiceClient targeting the given base URL (e.g. "http://localhost:8001").
func NewVoiceClient(baseURL string) *VoiceClient {
	return &VoiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Submit sends an async analysis request and returns the task submission details.
func (c *VoiceClient) Submit(data string, layers []string) (*VoiceSubmitResponse, error) {
	body, err := json.Marshal(VoiceAnalyzeRequest{Data: data, Layers: layers})
	if err != nil {
		return nil, fmt.Errorf("voice: marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/analyze/", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("voice: submit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		var errResp VoiceResultResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("voice: submit failed (%d): %s – %s", resp.StatusCode, errResp.Code, errResp.Message)
	}

	var out VoiceSubmitResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("voice: decode submit response: %w", err)
	}
	return &out, nil
}

// Poll fetches the current state of an async task.
func (c *VoiceClient) Poll(taskID string) (*VoiceResultResponse, int, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/analyze/" + taskID + "/")
	if err != nil {
		return nil, 0, fmt.Errorf("voice: poll: %w", err)
	}
	defer resp.Body.Close()

	var out VoiceResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("voice: decode poll response: %w", err)
	}
	return &out, resp.StatusCode, nil
}

// AnalyzeSync submits an analysis request in synchronous mode and waits for the result.
func (c *VoiceClient) AnalyzeSync(data string, layers []string) (*VoiceResultResponse, error) {
	body, err := json.Marshal(VoiceAnalyzeRequest{Data: data, Layers: layers})
	if err != nil {
		return nil, fmt.Errorf("voice: marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/analyze/?sync=true", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("voice: analyze sync: %w", err)
	}
	defer resp.Body.Close()

	var out VoiceResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("voice: decode sync response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voice: analyze sync failed (%d): %s – %s", resp.StatusCode, out.Code, out.Message)
	}
	return &out, nil
}
