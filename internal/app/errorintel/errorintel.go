package errorintel

import (
	"regexp"
	"strings"

	"github.com/rickseven/logiq/internal/domain"
)

// ErrorPattern defines a regex pattern and its metadata for detecting errors
type ErrorPattern struct {
	Regex     *regexp.Regexp
	ErrorType string
	Priority  int // 1: Build failure, 2: Exception, 3: Explicit ERROR, 4: Warning
	Formatter func(match []string) string
}

var patterns = []ErrorPattern{
	// Priority 1: Build failure markers
	{
		Regex:     regexp.MustCompile(`Module not found: Can't resolve '(.*?)'`),
		ErrorType: "module_resolution",
		Priority:  1,
		Formatter: func(match []string) string { return "Module not found '" + match[1] + "'" },
	},
	{
		Regex:     regexp.MustCompile(`Cannot find file '(.*?)'`),
		ErrorType: "file_not_found",
		Priority:  1,
		Formatter: func(match []string) string { return "Cannot find file '" + match[1] + "'" },
	},
	{
		Regex:     regexp.MustCompile(`(?i)build failed:?\s*(.*)`),
		ErrorType: "build_failure",
		Priority:  1,
		Formatter: func(match []string) string { return "Build failed: " + strings.TrimSpace(match[1]) },
	},

	// Priority 2: Exceptions and panics
	{
		Regex:     regexp.MustCompile(`(?i)Exception:\s*(.*)`),
		ErrorType: "exception",
		Priority:  2,
		Formatter: func(match []string) string { return "Exception: " + match[1] },
	},
	{
		Regex:     regexp.MustCompile(`(?i)panic:\s*(.*)`),
		ErrorType: "panic",
		Priority:  2,
		Formatter: func(match []string) string { return "Panic: " + match[1] },
	},
	{
		Regex:     regexp.MustCompile(`(?i)fatal error:\s*(.*)`),
		ErrorType: "fatal_error",
		Priority:  2,
		Formatter: func(match []string) string { return "Fatal error: " + match[1] },
	},

	// Priority 3: Explicit ERROR messages
	{
		Regex:     regexp.MustCompile(`\d+:\d+\s+error\s+(.*)`),
		ErrorType: "lint_error",
		Priority:  3,
		Formatter: func(match []string) string { return match[1] },
	},
	{
		Regex:     regexp.MustCompile(`(?i)^Error:\s*(.*)`),
		ErrorType: "generic_error",
		Priority:  3,
		Formatter: func(match []string) string { return match[1] },
	},
	{
		Regex:     regexp.MustCompile(`^ERROR\s+(.*)`),
		ErrorType: "generic_error",
		Priority:  3,
		Formatter: func(match []string) string { return match[1] },
	},

	// Priority 4: Warnings
	{
		Regex:     regexp.MustCompile(`(?i)^Warning:\s*(.*)`),
		ErrorType: "warning",
		Priority:  4,
		Formatter: func(match []string) string { return strings.TrimSpace(match[1]) },
	},
}

// Analyze extracts the most relevant error intel from the logs, correlating with code changes
func Analyze(logs []string, changedFiles []string) *domain.ErrorIntel {
	if len(logs) == 0 {
		return nil
	}

	bestMatchLineIdx := -1
	bestPriority := 999
	var bestCause string
	var bestType string

	for i, line := range logs {
		trimmed := strings.TrimSpace(line)
		for _, pat := range patterns {
			if pat.Priority >= bestPriority {
				continue
			}

			if match := pat.Regex.FindStringSubmatch(trimmed); len(match) > 0 {
				bestPriority = pat.Priority
				bestCause = pat.Formatter(match)
				bestType = pat.ErrorType
				bestMatchLineIdx = i
				if bestPriority == 1 {
					break
				}
			}
		}
	}

	if bestMatchLineIdx == -1 {
		return nil
	}

	// Correlation Logic
	var correlated []string
	if len(changedFiles) > 0 {
		// Look for any mention of changed files near the error
		searchRangeStart := bestMatchLineIdx - 5
		if searchRangeStart < 0 {
			searchRangeStart = 0
		}
		searchRangeEnd := bestMatchLineIdx + 5
		if searchRangeEnd > len(logs) {
			searchRangeEnd = len(logs)
		}

		foundFiles := make(map[string]bool)
		for i := searchRangeStart; i < searchRangeEnd; i++ {
			for _, cf := range changedFiles {
				if strings.Contains(logs[i], cf) {
					foundFiles[cf] = true
				}
			}
		}
		for f := range foundFiles {
			correlated = append(correlated, f)
		}
	}

	if len(correlated) > 0 {
		bestCause += "\n\nNote: This error might be related to recent changes in: " + strings.Join(correlated, ", ")
	}

	start := bestMatchLineIdx - 1
	if start < 0 {
		start = 0
	}
	end := bestMatchLineIdx + 2
	if end > len(logs) {
		end = len(logs)
	}

	contextLines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		contextLines = append(contextLines, logs[i])
	}

	return &domain.ErrorIntel{
		RootCause:         bestCause,
		ErrorType:         bestType,
		Context:           contextLines,
		CorrelatedChanges: correlated,
	}
}
