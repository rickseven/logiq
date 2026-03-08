package debugassist

import (
	"regexp"
	"strings"

	"github.com/rickseven/logiq/internal/domain"
)

var defaultRules = []domain.SuggestionRule{
	{
		Pattern:    regexp.MustCompile(`(?i)module not found|can't resolve`),
		ErrorType:  "module_resolution",
		Suggestion: "Check import path or ensure the component exists.",
	},
	{
		Pattern:    regexp.MustCompile(`(?i)cannot find file`),
		ErrorType:  "file_not_found",
		Suggestion: "Verify the file path and ensure it has not been renamed or deleted.",
	},
	{
		Pattern:    regexp.MustCompile(`(?i)missing semicolon`),
		ErrorType:  "lint_error",
		Suggestion: "Add a semicolon at the indicated line.",
	},
	{
		Pattern:    regexp.MustCompile(`(?i)gradle task failed`),
		ErrorType:  "build_failure",
		Suggestion: "Run: flutter clean && flutter pub get",
	},
	{
		Pattern:    regexp.MustCompile(`(?i)tests failed|test suite failed`),
		ErrorType:  "", // Matches across error types
		Suggestion: "Inspect failing test cases and verify expected outputs.",
	},
	{
		Pattern:    regexp.MustCompile(`(?i)address already in use|EADDRINUSE`),
		ErrorType:  "generic_error",
		Suggestion: "Port conflict detected. Try running: 'npx kill-port <port>' or use a different port.",
	},
	{
		Pattern:    regexp.MustCompile(`(?i)out of memory|heap limit`),
		ErrorType:  "",
		Suggestion: "Process ran out of memory. Try increasing Node.js memory limit with: 'NODE_OPTIONS=--max-old-space-size=4096'",
	},
	{
		Pattern:    regexp.MustCompile(`(?i)permission denied|EACCES`),
		ErrorType:  "",
		Suggestion: "Permission error. Check file ownership or try running with elevated privileges.",
	},
	{
		Pattern:    regexp.MustCompile(`(?i)ENOSPC|no space left on device`),
		ErrorType:  "",
		Suggestion: "Disk is full. Free up some space or check your temporary directory.",
	},
	{
		Pattern:    regexp.MustCompile(`(?i)unused variable`),
		ErrorType:  "warning",
		Suggestion: "Remove the unused variable or prefix it with an underscore if intentional.",
	},
}

// RegisterRule dynamically pushes rule extensions safely preserving order mapping
func RegisterRule(rule domain.SuggestionRule) {
	defaultRules = append([]domain.SuggestionRule{rule}, defaultRules...)
}

// Analyze matches explicit error intelligence patterns and generates suggestions
func Analyze(intel *domain.ErrorIntel) []string {
	if intel == nil || intel.RootCause == "" {
		return nil
	}

	var suggestions []string
	seen := make(map[string]bool)

	lowerCause := strings.ToLower(intel.RootCause)
	lowerType := strings.ToLower(intel.ErrorType)

	for _, rule := range defaultRules {
		// Filter by ErrorType if rule enforces one
		typeMatch := rule.ErrorType == "" || rule.ErrorType == lowerType
		if !typeMatch {
			continue
		}

		// Match root cause text
		if rule.Pattern.MatchString(lowerCause) {
			if !seen[rule.Suggestion] {
				suggestions = append(suggestions, rule.Suggestion)
				seen[rule.Suggestion] = true
			}
		}
	}

	return suggestions
}
