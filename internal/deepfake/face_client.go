package deepfake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
)

// DetectImage uploads an image to the face detection service and returns an instant result.
// The face detection service endpoint is POST /detect/image (multipart/form-data).
func (c *Client) DetectImage(file io.Reader, filename string) (*FaceImageResult, error) {
	body, contentType, err := buildMultipartFile(file, filename)
	if err != nil {
		return nil, fmt.Errorf("face: build request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/detect/image", contentType, body)
	if err != nil {
		return nil, fmt.Errorf("face: detect image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("face: detect image failed (%d): %s", resp.StatusCode, string(body))
	}

	var out FaceImageResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("face: decode image response: %w", err)
	}
	return &out, nil
}

// SubmitVideo uploads a video to the face detection service for async analysis.
// sampleEvery controls how many frames are skipped between analyses (0 uses the service default).
// The face detection service endpoint is POST /detect/video (multipart/form-data).
func (c *Client) SubmitVideo(file io.Reader, filename string, sampleEvery int) (*VideoSubmitResult, error) {
	body, contentType, err := buildMultipartFile(file, filename)
	if err != nil {
		return nil, fmt.Errorf("face: build video request: %w", err)
	}

	url := c.baseURL + "/detect/video"
	if sampleEvery > 0 {
		url += "?sample_every=" + strconv.Itoa(sampleEvery)
	}

	resp, err := c.httpClient.Post(url, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("face: submit video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("face: submit video failed (%d): %s", resp.StatusCode, string(body))
	}

	var out VideoSubmitResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("face: decode video submit response: %w", err)
	}
	return &out, nil
}

// PollVideo fetches the current status and result of a submitted video job.
// The face detection service endpoint is GET /detect/video/{job_id}.
func (c *Client) PollVideo(jobID string) (*VideoJobResult, int, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/detect/video/" + jobID)
	if err != nil {
		return nil, 0, fmt.Errorf("face: poll video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, resp.StatusCode, fmt.Errorf("face: poll video failed: %s", string(body))
	}

	var out VideoJobResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("face: decode video poll response: %w", err)
	}
	return &out, resp.StatusCode, nil
}

// ListVideoJobs returns all video analysis jobs tracked by the face detection service.
// The face detection service endpoint is GET /detect/video.
func (c *Client) ListVideoJobs() (*VideoJobListResult, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/detect/video")
	if err != nil {
		return nil, fmt.Errorf("face: list video jobs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("face: list video jobs failed (%d): %s", resp.StatusCode, string(body))
	}

	var out VideoJobListResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("face: decode video list response: %w", err)
	}
	return &out, nil
}

// buildMultipartFile wraps r into a multipart form body with field name "file".
func buildMultipartFile(r io.Reader, filename string) (*bytes.Buffer, string, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(part, r); err != nil {
		return nil, "", err
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &body, w.FormDataContentType(), nil
}
