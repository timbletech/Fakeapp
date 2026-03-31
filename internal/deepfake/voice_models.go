package deepfake

// VoiceAnalyzeRequest is the payload sent to the Go handler (converted from base64).
type VoiceAnalyzeRequest struct {
	Data   string   `json:"data"`
	Layers []string `json:"layers,omitempty"`
}

// VoiceSubmitResponse is returned by the voice service on async submission (202).
type VoiceSubmitResponse struct {
	Status          string   `json:"status"`
	ExecutionMode   string   `json:"execution_mode"`
	TaskID          string   `json:"task_id"`
	RequestedLayers []string `json:"requested_layers"`
	PollURL         string   `json:"poll_url"`
}

// VoiceResultResponse is the VARE API response from the voice analysis service.
// This matches the VARE (Voice Authenticity Risk Engine) API response schema.
type VoiceResultResponse struct {
	Status          string                `json:"status"`
	Filename        string                `json:"filename,omitempty"`
	DurationSec     float64               `json:"duration_sec,omitempty"`
	NSegments       int                   `json:"n_segments,omitempty"`
	EnsembleScore   float64               `json:"ensemble_score,omitempty"`
	RiskLevel       string                `json:"risk_level,omitempty"`
	RiskColor       string                `json:"risk_color,omitempty"`
	ModelConfidence float64               `json:"model_confidence,omitempty"`
	Models          map[string]VoiceModel `json:"models,omitempty"`
	Code            string                `json:"code,omitempty"`
	Message         string                `json:"message,omitempty"`
}

// VoiceModel represents a single model's analysis results.
type VoiceModel struct {
	Available          bool      `json:"available"`
	Error              *string   `json:"error"`
	AggregateSpoofProb *float64  `json:"aggregate_spoof_prob"`
	Verdict            *string   `json:"verdict"`
	SegmentScores      []float64 `json:"segment_scores"`
}

// Legacy models - kept for backward compatibility if needed
// ============================================================

// VoiceLayers holds per-layer voice analysis results.
type VoiceLayers struct {
	Metadata *VoiceMetadataLayer `json:"metadata,omitempty"`
	Spectral *VoiceSpectralLayer `json:"spectral,omitempty"`
	AIModel  *VoiceAIModelLayer  `json:"ai_model,omitempty"`
}

type VoiceMetadataLayer struct {
	Status         string               `json:"status"`
	RiskScore      float64              `json:"risk_score"`
	Classification string               `json:"classification"`
	AnomalyCount   int                  `json:"anomaly_count"`
	Report         *VoiceMetadataReport `json:"report,omitempty"`
}

type VoiceMetadataReport struct {
	FileName   string                 `json:"file_name"`
	Format     string                 `json:"format"`
	Duration   float64                `json:"duration_seconds"`
	SampleRate int                    `json:"sample_rate"`
	Channels   int                    `json:"channels"`
	Anomalies  []string               `json:"anomalies"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

type VoiceSpectralLayer struct {
	Status         string                `json:"status"`
	RiskScore      float64               `json:"risk_score"`
	Classification string                `json:"classification"`
	Metrics        *VoiceSpectralMetrics `json:"metrics,omitempty"`
}

type VoiceSpectralMetrics struct {
	SyntheticIndex float64 `json:"synthetic_index"`
	Correlation    float64 `json:"correlation"`
	NoiseEnergy    float64 `json:"noise_energy"`
	Threshold      float64 `json:"threshold"`
}

type VoiceAIModelLayer struct {
	Status                 string                       `json:"status"`
	RiskScore              float64                      `json:"risk_score"`
	Classification         string                       `json:"classification"`
	Components             *VoiceAIModelComponents      `json:"components,omitempty"`
	ComponentContributions []VoiceComponentContribution `json:"component_contributions,omitempty"`
}

type VoiceAIModelComponents struct {
	Classifier *VoiceClassifierResult `json:"classifier,omitempty"`
	LLM        *VoiceLLMResult        `json:"llm,omitempty"`
}

type VoiceClassifierResult struct {
	Status             string             `json:"status"`
	TamperProbability  float64            `json:"tamper_probability"`
	Classification     string             `json:"classification"`
	PredLabel          string             `json:"pred_label"`
	ClassProbabilities map[string]float64 `json:"class_probabilities"`
	ModelName          string             `json:"model_name"`
}

type VoiceLLMResult struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type VoiceComponentContribution struct {
	Component     string  `json:"component"`
	Score         float64 `json:"score"`
	Weight        float64 `json:"weight"`
	WeightedScore float64 `json:"weighted_score"`
}

type VoiceSummary struct {
	OverallRiskScore *float64                   `json:"overall_risk_score"`
	Decision         string                     `json:"decision"`
	RiskLevel        string                     `json:"risk_level"`
	SuccessfulLayers int                        `json:"successful_layers"`
	FailedLayers     []string                   `json:"failed_layers"`
	SkippedLayers    []string                   `json:"skipped_layers"`
	Contributions    []VoiceSummaryContribution `json:"contributions"`
}

type VoiceSummaryContribution struct {
	Layer         string  `json:"layer"`
	RiskScore     float64 `json:"risk_score"`
	Weight        float64 `json:"weight"`
	WeightedScore float64 `json:"weighted_score"`
}

type VoiceTimings struct {
	LayerDurationsMs map[string]int `json:"layer_durations_ms"`
	TotalMs          int            `json:"total_ms"`
}

type VoiceRuntimeConfig struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
