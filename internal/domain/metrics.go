package domain

// Metrics extracted from the execution
type Metrics struct {
	TestsPassed       int     `json:"tests_passed"`
	TestsFailed       int     `json:"tests_failed"`
	Warnings          int     `json:"warnings"`
	Errors            int     `json:"errors"`
	DurationSeconds   float64 `json:"duration_seconds"`
	MaxRAMMB          float64 `json:"max_ram_mb,omitempty"`
	AvgCPUPercent     float64 `json:"avg_cpu_percent,omitempty"`
	TestFiles         int     `json:"test_files,omitempty"`
	ModulesCompiled   int     `json:"modules_compiled,omitempty"`
	BundleSize        string  `json:"bundle_size,omitempty"`
	ArtifactSize      string  `json:"artifact_size,omitempty"`
	OriginalBytes     int     `json:"original_bytes,omitempty"`
	CompressedBytes   int     `json:"compressed_bytes,omitempty"`
	SavingsPercentage float64 `json:"savings_percentage,omitempty"`
}
