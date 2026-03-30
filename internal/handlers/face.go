package handlers

import (
	"net/http"
	"strconv"

	"device_only/internal/deepfake"

	"github.com/google/uuid"
)

// FaceHandler handles face deepfake detection endpoints.
type FaceHandler struct {
	client *deepfake.Client
}

// NewFaceHandler creates a FaceHandler backed by the given deepfake client.
func NewFaceHandler(client *deepfake.Client) *FaceHandler {
	return &FaceHandler{client: client}
}

// SetupRoutes registers face detection routes on mux.
func (h *FaceHandler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/face/image", h.handleDetectImage)
	mux.HandleFunc("POST /v1/face/video", h.handleSubmitVideo)
	mux.HandleFunc("GET /v1/face/video/{job_id}", h.handlePollVideo)
	mux.HandleFunc("GET /v1/face/video", h.handleListJobs)
}

// handleDetectImage handles POST /v1/face/image.
// Accepts multipart/form-data with a "file" field and returns an instant verdict.
func (h *FaceHandler) handleDetectImage(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"request_id": requestID,
			"error":      "invalid_request",
			"message":    "Failed to parse multipart form",
		})
		return
	}

	f, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"request_id": requestID,
			"error":      "missing_file",
			"message":    "No file uploaded. Use the 'file' field.",
		})
		return
	}
	defer f.Close()

	result, err := h.client.DetectImage(f, header.Filename)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"request_id": requestID,
			"error":      "upstream_error",
			"message":    err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"request_id": requestID,
		"status":     result.Status,
		"input":      result.Input,
		"result":     result.Result,
		"meta":       result.Meta,
	})
}

// handleSubmitVideo handles POST /v1/face/video.
// Accepts multipart/form-data with a "file" field. Supports ?sample_every=N query param.
// Returns a job_id to poll with GET /v1/face/video/{job_id}.
func (h *FaceHandler) handleSubmitVideo(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()

	if err := r.ParseMultipartForm(512 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"request_id": requestID,
			"error":      "invalid_request",
			"message":    "Failed to parse multipart form",
		})
		return
	}

	f, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"request_id": requestID,
			"error":      "missing_file",
			"message":    "No file uploaded. Use the 'file' field.",
		})
		return
	}
	defer f.Close()

	sampleEvery, _ := strconv.Atoi(r.URL.Query().Get("sample_every"))

	submitted, err := h.client.SubmitVideo(f, header.Filename, sampleEvery)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"request_id": requestID,
			"error":      "upstream_error",
			"message":    err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"request_id":   requestID,
		"status":       submitted.Status,
		"job_id":       submitted.JobID,
		"filename":     submitted.Filename,
		"sample_every": submitted.SampleEvery,
		"poll_url":     "/v1/face/video/" + submitted.JobID,
	})
}

// handlePollVideo handles GET /v1/face/video/{job_id}.
// Returns the current status and result of a submitted video job.
func (h *FaceHandler) handlePollVideo(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()
	jobID := r.PathValue("job_id")

	if jobID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"request_id": requestID,
			"error":      "missing_job_id",
			"message":    "job_id is required",
		})
		return
	}

	result, statusCode, err := h.client.PollVideo(jobID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"request_id": requestID,
			"error":      "upstream_error",
			"message":    err.Error(),
		})
		return
	}

	writeJSON(w, statusCode, map[string]interface{}{
		"request_id": requestID,
		"job_id":     result.JobID,
		"filename":   result.Filename,
		"status":     result.Status,
		"progress":   result.Progress,
		"input":      result.Input,
		"result":     result.Result,
		"meta":       result.Meta,
		"error":      result.Error,
	})
}

// handleListJobs handles GET /v1/face/video.
// Returns all video analysis jobs tracked by the face detection service.
func (h *FaceHandler) handleListJobs(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()

	jobs, err := h.client.ListVideoJobs()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"request_id": requestID,
			"error":      "upstream_error",
			"message":    err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"request_id": requestID,
		"total":      jobs.Total,
		"jobs":       jobs.Jobs,
	})
}
