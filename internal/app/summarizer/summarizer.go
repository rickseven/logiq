package summarizer

import (
	"regexp"
	"strings"
	"time"

	"github.com/rickseven/logiq/internal/domain"
)

var (
	zeroErrorRegex = regexp.MustCompile(`0\s+errors?`)
	errorPattern   = regexp.MustCompile(`(?i)(?:ERROR|FAILED|PANIC)`)
)

// Summarize acts as the semantic log summarizer
func Summarize(compressedLogs []string, duration time.Duration) domain.Summary {
	var errors, warnings int
	status := "success"
	important := []string{}

	isBuildSucceeded := false
	isError := false
	isWarning := false
	isTestCompleted := false
	isBuildCompleted := false

	for _, l := range compressedLogs {
		text := strings.ToLower(l)

		// Check for errors, but ignore "0 errors" or "0 failures"
		if errorPattern.MatchString(l) {
			// Specific exclusion for "0 errors", "0 failures"
			if zeroErrorRegex.MatchString(text) || strings.Contains(text, "0 failures") || strings.Contains(text, "0 failed") {
				continue
			}

			errors++
			important = append(important, l)
			if strings.Contains(text, "error") && !zeroErrorRegex.MatchString(text) {
				isError = true
			}
		}

		// Check for warnings
		if strings.Contains(text, "warning") || strings.Contains(text, "warn") {
			warnings++
			if strings.Contains(text, "warning") || strings.Contains(text, "warn") {
				isWarning = true
			}
		}

		// Success signals
		if strings.Contains(text, "build succeeded") || strings.Contains(text, "compiled successfully") {
			isBuildSucceeded = true
			isBuildCompleted = true
		} else if strings.Contains(text, "build complete") || strings.Contains(text, "build finished") || strings.Contains(text, "dist/") {
			isBuildCompleted = true
		}

		if strings.Contains(text, "test") && (strings.Contains(text, "completed") || strings.Contains(text, "passed") || strings.Contains(text, "finished")) {
			isTestCompleted = true
		}
	}

	if errors > 0 {
		status = "failure"
	}

	if len(important) > 10 {
		important = important[:10]
	}

	var summaryText string
	if isError {
		if isBuildCompleted {
			summaryText = "\u2717 Completed with errors"
		} else {
			summaryText = "\u2717 Build failed"
		}
	} else if errors > 0 {
		summaryText = "\u2717 Execution failed"
	} else if isBuildSucceeded {
		summaryText = "\u2713 Build succeeded"
	} else if isBuildCompleted {
		summaryText = "\u2713 Build completed"
	} else if isTestCompleted {
		summaryText = "\u2713 Tests completed"
	} else if isWarning || warnings > 0 {
		summaryText = "\u26A0 Completed with warnings"
	} else {
		summaryText = "\u2713 Execution successful"
	}

	return domain.Summary{
		Status:          status,
		SummaryText:     summaryText,
		ImportantEvents: important,
		Metrics: domain.Metrics{
			Errors:          errors,
			Warnings:        warnings,
			DurationSeconds: duration.Seconds(),
		},
	}
}
