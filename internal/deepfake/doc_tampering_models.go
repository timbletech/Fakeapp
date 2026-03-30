package deepfake

// AnalyzeRequest is the payload sent to the deepfake service.
type AnalyzeRequest struct {
	Data   string   `json:"data"`
	Layers []string `json:"layers,omitempty"`
}

// SubmitResponse is returned by the deepfake service on async submission (202).
type SubmitResponse struct {
	Status          string   `json:"status"`
	ExecutionMode   string   `json:"execution_mode"`
	TaskID          string   `json:"task_id"`
	RequestedLayers []string `json:"requested_layers"`
	PollURL         string   `json:"poll_url"`
}

// ResultResponse is returned when a task is complete (200) or pending (202 on poll).
type ResultResponse struct {
	Status          string         `json:"status"`
	ExecutionMode   string         `json:"execution_mode"`
	TaskID          string         `json:"task_id,omitempty"`
	TaskState       string         `json:"task_state,omitempty"`
	RequestedLayers []string       `json:"requested_layers,omitempty"`
	Layers          *Layers        `json:"layers,omitempty"`
	Summary         *Summary       `json:"summary,omitempty"`
	Timings         *Timings       `json:"timings,omitempty"`
	RuntimeConfig   *RuntimeConfig `json:"runtime_config,omitempty"`
	// Error fields (returned by deepfake service on 400/500)
	Code            string   `json:"code,omitempty"`
	Message         string   `json:"message,omitempty"`
	SupportedLayers []string `json:"supported_layers,omitempty"`
}

// Layers holds per-layer analysis results.
type Layers struct {
	Metadata *MetadataLayer `json:"metadata,omitempty"`
	VAE      *VAELayer      `json:"vae,omitempty"`
	AIModel  *AIModelLayer  `json:"ai_model,omitempty"`
}

type MetadataLayer struct {
	Status         string          `json:"status"`
	RiskScore      float64         `json:"risk_score"`
	Classification string          `json:"classification"`
	AnomalyCount   int             `json:"anomaly_count"`
	Report         *MetadataReport `json:"report,omitempty"`
}

type MetadataReport struct {
	FileName   string                 `json:"file_name"`
	Format     string                 `json:"format"`
	Dimensions string                 `json:"dimensions"`
	EXIF       map[string]interface{} `json:"exif"`
	Anomalies  []string               `json:"anomalies"`
}

type VAELayer struct {
	Status         string      `json:"status"`
	RiskScore      float64     `json:"risk_score"`
	Classification string      `json:"classification"`
	Metrics        *VAEMetrics `json:"metrics,omitempty"`
}

type VAEMetrics struct {
	SyntheticIndex float64 `json:"synthetic_index"`
	Correlation    float64 `json:"correlation"`
	NoiseEnergy    float64 `json:"noise_energy"`
	Threshold      float64 `json:"threshold"`
	Device         string  `json:"device"`
}

type AIModelLayer struct {
	Status                 string                  `json:"status"`
	RiskScore              float64                 `json:"risk_score"`
	Classification         string                  `json:"classification"`
	Components             *AIModelComponents      `json:"components,omitempty"`
	ComponentContributions []ComponentContribution `json:"component_contributions,omitempty"`
}

type AIModelComponents struct {
	ResNet18 *ResnetResult `json:"resnet18,omitempty"`
	LLM      *LLMResult    `json:"llm,omitempty"`
}

type ResnetResult struct {
	Status             string             `json:"status"`
	TamperProbability  float64            `json:"tamper_probability"`
	Classification     string             `json:"classification"`
	PredLabel          string             `json:"pred_label"`
	ClassProbabilities map[string]float64 `json:"class_probabilities"`
	Device             string             `json:"device"`
	ModelName          string             `json:"model_name"`
	ImageSize          int                `json:"image_size"`
}

type LLMResult struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type ComponentContribution struct {
	Component     string  `json:"component"`
	Score         float64 `json:"score"`
	Weight        float64 `json:"weight"`
	WeightedScore float64 `json:"weighted_score"`
}

type Summary struct {
	OverallRiskScore *float64              `json:"overall_risk_score"`
	Decision         string                `json:"decision"`
	RiskLevel        string                `json:"risk_level"`
	SuccessfulLayers int                   `json:"successful_layers"`
	FailedLayers     []string              `json:"failed_layers"`
	SkippedLayers    []string              `json:"skipped_layers"`
	Contributions    []SummaryContribution `json:"contributions"`
}

type SummaryContribution struct {
	Layer         string  `json:"layer"`
	RiskScore     float64 `json:"risk_score"`
	Weight        float64 `json:"weight"`
	WeightedScore float64 `json:"weighted_score"`
}

type Timings struct {
	LayerDurationsMs map[string]int `json:"layer_durations_ms"`
	TotalMs          int            `json:"total_ms"`
}

type RuntimeConfig struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
