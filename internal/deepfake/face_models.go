package deepfake

// FaceInput describes the input type returned by the face detection service.
type FaceInput struct {
	Type string `json:"type"`
}

// FaceResult holds the verdict and confidence from the face detection service.
type FaceResult struct {
	Verdict         string  `json:"verdict"`
	Confidence      string  `json:"confidence"`
	FakeProbability float64 `json:"fake_probability,omitempty"`
}

// FaceMeta holds face count metadata from the face detection service.
type FaceMeta struct {
	FacesDetected int     `json:"faces_detected,omitempty"`
	MultipleFaces bool    `json:"multiple_faces,omitempty"`
	DurationSec   float64 `json:"duration_sec,omitempty"`
	FPS           float64 `json:"fps,omitempty"`
	TotalFrames   int     `json:"total_frames,omitempty"`
	Width         int     `json:"width,omitempty"`
	Height        int     `json:"height,omitempty"`
	FileSize      int64   `json:"file_size,omitempty"`
}

// FaceImageResult is returned by POST /detect/image on the face detection service.
type FaceImageResult struct {
	Status string     `json:"status"`
	Input  FaceInput  `json:"input"`
	Result FaceResult `json:"result"`
	Meta   FaceMeta   `json:"meta"`
	// Detail is populated on error responses (400/500).
	Detail string `json:"detail,omitempty"`
}

// VideoSubmitResult is returned by POST /detect/video.
// When the upstream returns a synchronous result, JobID is empty and Input/Result/Meta are populated.
type VideoSubmitResult struct {
	Status      string      `json:"status"`
	JobID       string      `json:"job_id,omitempty"`
	Filename    string      `json:"filename,omitempty"`
	SampleEvery int         `json:"sample_every,omitempty"`
	PollURL     string      `json:"poll_url,omitempty"`
	Input       *FaceInput  `json:"input,omitempty"`
	Result      *FaceResult `json:"result,omitempty"`
	Meta        *FaceMeta   `json:"meta,omitempty"`
	// Detail is populated on error responses.
	Detail string `json:"detail,omitempty"`
}

// VideoJobResult is returned by GET /detect/video/{job_id}.
type VideoJobResult struct {
	JobID    string      `json:"job_id"`
	Filename string      `json:"filename"`
	Status   string      `json:"status"`
	Progress float64     `json:"progress"`
	Input    *FaceInput  `json:"input,omitempty"`
	Result   *FaceResult `json:"result,omitempty"`
	Meta     *FaceMeta   `json:"meta,omitempty"`
	// Error is set when Status == "error".
	Error string `json:"error,omitempty"`
	// Detail is populated on 404 responses.
	Detail string `json:"detail,omitempty"`
}

// VideoJobSummary is a single entry in the job list.
type VideoJobSummary struct {
	JobID    string  `json:"job_id"`
	Filename string  `json:"filename"`
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
}

// VideoJobListResult is returned by GET /detect/video.
type VideoJobListResult struct {
	Total int               `json:"total"`
	Jobs  []VideoJobSummary `json:"jobs"`
}
