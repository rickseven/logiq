package domain

// Summary represents the semantic meaning extracted from logs
type Summary struct {
	Status          string      `json:"status"`  // "success" or "failure"
	SummaryText     string      `json:"summary"` // Short summary optimized for LLM
	ImportantEvents []string    `json:"important_events,omitempty"`
	Metrics         Metrics     `json:"metrics"`
	ErrorIntel      *ErrorIntel `json:"error_intel,omitempty"`
	Suggestions     []string    `json:"suggestions,omitempty"`
}
