package domain

// ErrorIntel provides root cause analysis for failures
type ErrorIntel struct {
	RootCause         string   `json:"root_cause"`
	ErrorType         string   `json:"error_type"`
	Context           []string `json:"context"`
	CorrelatedChanges []string `json:"correlated_changes,omitempty"`
}
