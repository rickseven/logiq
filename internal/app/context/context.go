package ctxengine

import (
	"fmt"
	"github.com/rickseven/logiq/internal/domain"
	"strings"
)

// Compressor represents the AI Context Compression engine
type Compressor struct {
	logs []string
}

// NewCompressor initializes a logger sequence
func NewCompressor(logs []string) *Compressor {
	return &Compressor{logs: logs}
}

// Stage1FilterNoise removes noise like progress indicators and downloads
func (c *Compressor) Stage1FilterNoise() []string {
	var filtered []string
	for _, line := range c.logs {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "progress") || strings.Contains(lower, "downloading") || strings.Contains(lower, "fetching") {
			continue // skip noise
		}
		filtered = append(filtered, line)
	}
	return filtered
}

// Stage2ExtractKeyEvents extracts the most critical log lines
func (c *Compressor) Stage2ExtractKeyEvents(filtered []string) []string {
	var events []string
	for _, line := range filtered {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") ||
			strings.Contains(lower, "fail") ||
			strings.Contains(lower, "success") ||
			strings.Contains(lower, "build") ||
			strings.Contains(lower, "test") ||
			strings.Contains(lower, "built") ||
			strings.Contains(lower, "artifact") {
			events = append(events, strings.TrimSpace(line))
		}
	}

	// Enforce strict Token Budget (roughly limit lines returned)
	if len(events) > 15 {
		// Just keep the last 15 most important ones to stay well under 100-200 tokens
		events = events[len(events)-15:]
	}

	return events
}

// Compress executes all 4 stages and returns the structured block snippet
func (c *Compressor) Compress(out *domain.StructuredOutput) string {
	// Stage 1
	filtered := c.Stage1FilterNoise()

	// Stage 2
	keyEvents := c.Stage2ExtractKeyEvents(filtered)

	// Stage 3: Metric Extraction (Already embedded safely within out.Metrics via parsing)
	// We read directly from out.Metrics for consistency

	// Stage 4: Structured Context
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Command: %s\n", out.Command))
	b.WriteString(fmt.Sprintf("Status: %s\n", out.Status))

	if out.Metrics.ModulesCompiled > 0 {
		b.WriteString(fmt.Sprintf("Modules compiled: %d\n", out.Metrics.ModulesCompiled))
	}
	if out.Metrics.TestsPassed > 0 || out.Metrics.TestsFailed > 0 {
		if out.Metrics.TestsPassed > 0 {
			b.WriteString(fmt.Sprintf("Tests passed: %d\n", out.Metrics.TestsPassed))
		}
		if out.Metrics.TestsFailed > 0 {
			b.WriteString(fmt.Sprintf("Tests failed: %d\n", out.Metrics.TestsFailed))
		}
	}
	if out.Metrics.BundleSize != "" {
		b.WriteString(fmt.Sprintf("Bundle size: %s\n", out.Metrics.BundleSize))
	}
	if out.Metrics.ArtifactSize != "" {
		b.WriteString(fmt.Sprintf("Artifact size: %s\n", out.Metrics.ArtifactSize))
	}
	if out.Metrics.DurationSeconds > 0 {
		b.WriteString(fmt.Sprintf("Duration: %.2fs\n", out.Metrics.DurationSeconds))
	}

	if len(keyEvents) > 0 && out.Status != "success" {
		b.WriteString("\nKey Events:\n")
		for _, e := range keyEvents {
			b.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	return strings.TrimSpace(b.String())
}
