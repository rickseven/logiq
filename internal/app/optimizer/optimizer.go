package optimizer

import (
	"fmt"
	"regexp"
	"strings"
)

// Optimizer reduces the tokens of raw logs for AI consumption
type Optimizer struct {
	passRegex *regexp.Regexp
	failRegex *regexp.Regexp
}

func NewOptimizer() *Optimizer {
	return &Optimizer{
		// Common test frameworks PASS/FAIL formats
		passRegex: regexp.MustCompile(`^(?i)(?:\x{2713}|PASS|ok)\s+(.*)`),
		failRegex: regexp.MustCompile(`^(?i)(?:\x{2717}|FAIL)\s+(.*)`),
	}
}

// Optimize reduces a batch of lines by grouping and filtering
func (o *Optimizer) Optimize(logs []string) []string {
	if len(logs) == 0 {
		return logs
	}

	var optimized []string
	var lastLine string

	passCount := 0
	failCount := 0

	// Flush grouped events
	flush := func() {
		if passCount > 0 {
			optimized = append(optimized, fmt.Sprintf("\u2713 %d tests passed", passCount))
			passCount = 0
		}
		if failCount > 0 {
			optimized = append(optimized, fmt.Sprintf("\u2717 %d tests failed", failCount))
			failCount = 0
		}
	}

	for _, line := range logs {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Noise filtering
		if isNoise(trimmed) {
			continue
		}

		// Duplicate removal
		if trimmed == lastLine {
			continue
		}
		lastLine = trimmed

		// Event Grouping
		if o.passRegex.MatchString(trimmed) {
			passCount++
			continue
		}
		if o.failRegex.MatchString(trimmed) {
			failCount++
			continue
		}

		// If we encounter a normal line, flush the grouped counts first
		if passCount > 0 || failCount > 0 {
			flush()
		}

		optimized = append(optimized, trimmed)
	}

	flush()

	return optimized
}

func isNoise(line string) bool {
	lower := strings.ToLower(line)
	// Example noise patterns
	if strings.Contains(lower, "node_modules") && strings.Contains(lower, "warning") {
		return true // Example: suppress warnings from dependencies
	}
	if strings.HasPrefix(lower, "[info] downloading") || strings.HasPrefix(lower, "[info] fetching") {
		return true
	}
	return false
}
