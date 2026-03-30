package handlers

import (
	"encoding/json"
	"net/http"

	"device_only/internal/deepfake"

	"github.com/google/uuid"
)

// VoiceAnalyzeRequest is the payload accepted by the voice analysis endpoints.
type VoiceAnalyzeRequest struct {
	Data   string   `json:"data"`
	Layers []string `json:"layers,omitempty"`
}

// VoiceSubmitResponse wraps the voice service's async submission response.
type VoiceSubmitResponse struct {
	RequestID       string   `json:"request_id"`
	Status          string   `json:"status"`
	ExecutionMode   string   `json:"execution_mode"`
	TaskID          string   `json:"task_id"`
	RequestedLayers []string `json:"requested_layers"`
	PollURL         string   `json:"poll_url"`
}

// VoiceResultResponse wraps the voice service's result (sync or polled).
type VoiceResultResponse struct {
	RequestID string `json:"request_id"`
	*deepfake.VoiceResultResponse
}

// VoiceHandler handles voice analysis endpoints.
type VoiceHandler struct {
	client *deepfake.VoiceClient
}

// NewVoiceHandler creates a VoiceHandler backed by the given client.
func NewVoiceHandler(client *deepfake.VoiceClient) *VoiceHandler {
	return &VoiceHandler{client: client}
}

// SetupRoutes registers voice routes on mux.
func (h *VoiceHandler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/voice/analyze", h.handleAnalyze)
	mux.HandleFunc("GET /v1/voice/analyze/{task_id}", h.handlePoll)
}

// handleAnalyze handles POST /v1/voice/analyze[?sync=true].
//
//   - Default (async): submits the audio to the voice service and returns a
//     task_id the caller can poll with GET /v1/voice/analyze/{task_id}.
//   - Sync (?sync=true): waits for the voice service to finish and returns
//     the full result in a single call.
func (h *VoiceHandler) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()

	var req VoiceAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"request_id": requestID,
			"error":      "invalid_request",
			"message":    "Invalid JSON in request body",
		})
		return
	}

	if req.Data == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"request_id": requestID,
			"error":      "missing_data",
			"message":    "No base64 audio data provided. Use the 'data' field.",
		})
		return
	}

	if r.URL.Query().Get("sync") == "true" {
		result, err := h.client.AnalyzeSync(req.Data, req.Layers)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{
				"request_id": requestID,
				"error":      "upstream_error",
				"message":    err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, VoiceResultResponse{RequestID: requestID, VoiceResultResponse: result})
		return
	}

	submitted, err := h.client.Submit(req.Data, req.Layers)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"request_id": requestID,
			"error":      "upstream_error",
			"message":    err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusAccepted, VoiceSubmitResponse{
		RequestID:       requestID,
		Status:          submitted.Status,
		ExecutionMode:   submitted.ExecutionMode,
		TaskID:          submitted.TaskID,
		RequestedLayers: submitted.RequestedLayers,
		PollURL:         "/v1/voice/analyze/" + submitted.TaskID,
	})
}

// handlePoll handles GET /v1/voice/analyze/{task_id}.
// It proxies the poll request to the voice service and forwards the result.
func (h *VoiceHandler) handlePoll(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	taskID := r.PathValue("task_id")

	if taskID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"request_id": requestID,
			"error":      "missing_task_id",
			"message":    "task_id is required",
		})
		return
	}

	result, statusCode, err := h.client.Poll(taskID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"request_id": requestID,
			"error":      "upstream_error",
			"message":    err.Error(),
		})
		return
	}

	writeJSON(w, statusCode, VoiceResultResponse{RequestID: requestID, VoiceResultResponse: result})
}
