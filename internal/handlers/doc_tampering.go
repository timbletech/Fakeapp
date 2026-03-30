package handlers

import (
	"encoding/json"
	"net/http"

	"device_only/internal/deepfake"

	"github.com/google/uuid"
)

// DeepfakeAnalyzeRequest is the payload accepted by the Go API's deepfake endpoints.
type DeepfakeAnalyzeRequest struct {
	Data   string   `json:"data"`
	Layers []string `json:"layers,omitempty"`
}

// DeepfakeSubmitResponse wraps the deepfake service's async submission response.
type DeepfakeSubmitResponse struct {
	RequestID       string   `json:"request_id"`
	Status          string   `json:"status"`
	ExecutionMode   string   `json:"execution_mode"`
	TaskID          string   `json:"task_id"`
	RequestedLayers []string `json:"requested_layers"`
	PollURL         string   `json:"poll_url"`
}

// DeepfakeResultResponse wraps the deepfake service's result (sync or polled).
type DeepfakeResultResponse struct {
	RequestID string `json:"request_id"`
	*deepfake.ResultResponse
}

// DeepfakeHandler handles deepfake analysis endpoints.
type DeepfakeHandler struct {
	client *deepfake.Client
}

// NewDeepfakeHandler creates a DeepfakeHandler backed by the given client.
func NewDeepfakeHandler(client *deepfake.Client) *DeepfakeHandler {
	return &DeepfakeHandler{client: client}
}

// SetupRoutes registers deepfake routes on mux.
func (h *DeepfakeHandler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/deepfake/analyze", h.handleAnalyze)
	mux.HandleFunc("GET /v1/deepfake/analyze/{task_id}", h.handlePoll)
}

// handleAnalyze handles POST /v1/deepfake/analyze[?sync=true].
//
//   - Default (async): submits the image to the deepfake service and returns a
//     task_id the caller can poll with GET /v1/deepfake/analyze/{task_id}.
//   - Sync (?sync=true): waits for the deepfake service to finish and returns
//     the full result in a single call.
func (h *DeepfakeHandler) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()

	var req DeepfakeAnalyzeRequest
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
			"message":    "No base64 image data provided. Use the 'data' field.",
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
		writeJSON(w, http.StatusOK, DeepfakeResultResponse{RequestID: requestID, ResultResponse: result})
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

	writeJSON(w, http.StatusAccepted, DeepfakeSubmitResponse{
		RequestID:       requestID,
		Status:          submitted.Status,
		ExecutionMode:   submitted.ExecutionMode,
		TaskID:          submitted.TaskID,
		RequestedLayers: submitted.RequestedLayers,
		PollURL:         "/v1/deepfake/analyze/" + submitted.TaskID,
	})
}

// handlePoll handles GET /v1/deepfake/analyze/{task_id}.
// It proxies the poll request to the deepfake service and forwards the result.
func (h *DeepfakeHandler) handlePoll(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, statusCode, DeepfakeResultResponse{RequestID: requestID, ResultResponse: result})
}
