package domain

// ExplainResult represents the structured explanation of a command
type ExplainResult struct {
	Command     string `json:"command"`
	Type        string `json:"type"`
	Tool        string `json:"tool"`
	Description string `json:"description"`
}

// DoctorResult represents the environment diagnosis
type DoctorResult struct {
	Node         string   `json:"node"`
	Npm          string   `json:"npm"`
	Pnpm         string   `json:"pnpm"`
	Vite         string   `json:"vite"`
	Flutter      string   `json:"flutter"`
	Dart         string   `json:"dart"`
	Git          string   `json:"git"`
	ProjectType  string   `json:"project_type"`
	Capabilities []string `json:"capabilities"`
	Limitations  []string `json:"limitations"`
}

// TraceEntry represents an execution locally tracked
type TraceEntry struct {
	ExecutionID string `json:"execution_id,omitempty"`
	Command     string `json:"command"`
	Status      string `json:"status"`
	Summary     string `json:"summary,omitempty"`
}
