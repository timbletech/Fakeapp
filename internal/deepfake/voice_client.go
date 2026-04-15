package deepfake

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// VoiceClient wraps the voice analysis service HTTP API (VARE).
// API expects multipart/form-data with raw audio file, not JSON.
type VoiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewVoiceClient creates a VoiceClient targeting the given base URL (e.g. "http://localhost:8096").
func NewVoiceClient(baseURL string) *VoiceClient {
	return &VoiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Voice analysis can take time
		},
	}
}

// Submit sends an async analysis request and returns the task submission details.
// Note: VARE API doesn't support async, so this currently calls AnalyzeSync.
func (c *VoiceClient) Submit(data string, layers []string) (*VoiceSubmitResponse, error) {
	// VARE API only supports synchronous analysis, not async submission
	_, err := c.AnalyzeSync(data, layers)
	if err != nil {
		return nil, err
	}
	// Map sync result to submit response format
	return &VoiceSubmitResponse{
		Status: "ok",
	}, nil
}

// Poll fetches the current state of an async task.
// Note: VARE API doesn't support polling.
func (c *VoiceClient) Poll(taskID string) (*VoiceResultResponse, int, error) {
	return nil, 0, fmt.Errorf("voice: VARE API does not support polling")
}

// AnalyzeSync decodes base64 audio and sends it to VARE's /analyze endpoint as multipart/form-data.
// The base64 data parameter should contain audio data (e.g., from file upload).
func (c *VoiceClient) AnalyzeSync(data string, layers []string) (*VoiceResultResponse, error) {
	// Decode base64 to binary audio data
	audioData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("voice: invalid base64 data: %w", err)
	}

	// Create multipart form with audio file
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	// Add file field (VARE expects field name "file")
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, fmt.Errorf("voice: create form file: %w", err)
	}

	if _, err := io.Copy(part, bytes.NewReader(audioData)); err != nil {
		return nil, fmt.Errorf("voice: write audio data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("voice: close multipart writer: %w", err)
	}

	// Send request to VARE /analyze endpoint
	url := c.baseURL + "/analyze"
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return nil, fmt.Errorf("voice: create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("voice: analyze sync: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("voice: analyze failed (%d): %s", resp.StatusCode, string(body))
	}

	var out VoiceResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("voice: decode response: %w", err)
	}

	return &out, nil
}
