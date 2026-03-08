package compress

import (
	"fmt"
	"strings"
)

// Compress Logs reduces raw logs using an Ultra-Compact deduplication algorithm
// followed by truncation if it still exceeds limits.
func Compress(logs []string) []string {
	if len(logs) == 0 {
		return logs
	}

	// 1. Ultra-Compact Step: Run-Length Encoding / Deduplication
	var deduped []string
	var lastLine string
	var repetitionCount int

	for _, line := range logs {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue // Skip noisy empty lines in compact mode
		}

		if trimmed == lastLine {
			repetitionCount++
		} else {
			if repetitionCount > 0 {
				deduped = append(deduped, fmt.Sprintf("%s (Repeated %d times)", lastLine, repetitionCount+1))
			} else if lastLine != "" {
				deduped = append(deduped, lastLine)
			}
			lastLine = trimmed
			repetitionCount = 0
		}
	}
	// flush last pending
	if repetitionCount > 0 {
		deduped = append(deduped, fmt.Sprintf("%s (Repeated %d times)", lastLine, repetitionCount+1))
	} else if lastLine != "" {
		deduped = append(deduped, lastLine)
	}

	// 2. Truncation Step (Head / Tail)
	const maxLines = 800 // Increased from 200 since we've already cleaned noise
	if len(deduped) <= maxLines {
		return deduped
	}

	// Keep the first 300 and the last 500 (usually errors are at the tail)
	head := deduped[:300]
	tail := deduped[len(deduped)-500:]

	compressed := append([]string{}, head...)
	compressed = append(compressed, "\n... [TRUNCATED ULTRA-COMPACT BY LOGIQ: LOGS TOO LONG] ...\n")
	compressed = append(compressed, tail...)

	return compressed
}
