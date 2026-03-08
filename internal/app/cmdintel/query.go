package cmdintel

import (
	"strings"

	"github.com/rickseven/logiq/internal/domain"
)

// QueryHistory searches the execution history using keyword matching
func QueryHistory(query string) []domain.TraceEntry {
	entries := GetTraceEntries()
	if entries == nil {
		return nil
	}

	keywords := strings.Fields(strings.ToLower(query))
	var results []domain.TraceEntry

	for _, entry := range entries {
		cmdLower := strings.ToLower(entry.Command)
		summaryLower := strings.ToLower(entry.Summary)
		statusLower := strings.ToLower(entry.Status)

		matchCount := 0
		for _, kw := range keywords {
			if strings.Contains(cmdLower, kw) || strings.Contains(summaryLower, kw) || strings.Contains(statusLower, kw) {
				matchCount++
			}
		}

		// Basic "relevancy": if many keywords match, it's a good result
		if matchCount > 0 {
			results = append(results, entry)
		}
	}

	return results
}
