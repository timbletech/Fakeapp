package deepfake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client wraps the deepfake detection service HTTP API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Client targeting the given base URL (e.g. "http://localhost:8000").
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Submit sends an async analysis request and returns the task submission details.
func (c *Client) Submit(data string, layers []string) (*SubmitResponse, error) {
	body, err := json.Marshal(AnalyzeRequest{Data: data, Layers: layers})
	if err != nil {
		return nil, fmt.Errorf("deepfake: marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/analyze/", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deepfake: submit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		var errResp ResultResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("deepfake: submit failed (%d): %s – %s", resp.StatusCode, errResp.Code, errResp.Message)
	}

	var out SubmitResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("deepfake: decode submit response: %w", err)
	}
	return &out, nil
}

// Poll fetches the current state of an async task.
// Returns the result; callers should check result.Status ("pending" vs "success"/"error").
func (c *Client) Poll(taskID string) (*ResultResponse, int, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/analyze/" + taskID + "/")
	if err != nil {
		return nil, 0, fmt.Errorf("deepfake: poll: %w", err)
	}
	defer resp.Body.Close()

	var out ResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("deepfake: decode poll response: %w", err)
	}
	return &out, resp.StatusCode, nil
}

// AnalyzeSync submits an analysis request in synchronous mode and waits for the result.
func (c *Client) AnalyzeSync(data string, layers []string) (*ResultResponse, error) {
	body, err := json.Marshal(AnalyzeRequest{Data: data, Layers: layers})
	if err != nil {
		return nil, fmt.Errorf("deepfake: marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/analyze/?sync=true", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deepfake: analyze sync: %w", err)
	}
	defer resp.Body.Close()

	var out ResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("deepfake: decode sync response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deepfake: analyze sync failed (%d): %s – %s", resp.StatusCode, out.Code, out.Message)
	}
	return &out, nil
}

// Health checks the deepfake service health endpoint.
func (c *Client) Health() (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/health/")
	if err != nil {
		return nil, fmt.Errorf("deepfake: health: %w", err)
	}
	defer resp.Body.Close()

	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("deepfake: decode health response: %w", err)
	}
	return out, nil
}
