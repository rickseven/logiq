package domain

// StructuredOutput is the final machine readable JSON schema output
type StructuredOutput struct {
	ExecutionID       string      `json:"execution_id"`
	Tool              string      `json:"tool"`
	Command           string      `json:"command"`
	Status            string      `json:"status"`
	Summary           string      `json:"summary"`
	ImportantEvents   []string    `json:"important_events,omitempty"`
	Metrics           Metrics     `json:"metrics"`
	ErrorIntel        *ErrorIntel `json:"error_intel,omitempty"`
	Suggestions       []string    `json:"suggestions,omitempty"`
	ArtifactPath      string      `json:"artifact_path,omitempty"`
	RawLogPath        string      `json:"raw_log_path,omitempty"`
	CompressedContext string      `json:"compressed_context,omitempty"`
	Timestamp         string      `json:"timestamp"`
}
