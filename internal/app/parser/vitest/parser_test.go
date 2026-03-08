package vitest

import (
	"testing"
)

func TestVitestParser(t *testing.T) {
	parser := NewParser()
	lines := []string{
		"\u2713 components/Button.spec.ts (3 tests)",
		"\u2713 utils/helpers.spec.ts (5 tests)",
		"Test Files 2 passed (2)",
		"Tests 8 passed (8)",
	}

	for _, line := range lines {
		parser.Parse(line)
	}

	summary := parser.Summary()

	if summary.Status != "success" {
		t.Errorf("Expected success, got %s", summary.Status)
	}
	if summary.Metrics.TestFiles != 2 {
		t.Errorf("Expected 2 test files, got %d", summary.Metrics.TestFiles)
	}
	if summary.Metrics.TestsPassed != 8 {
		t.Errorf("Expected 8 tests passed, got %d", summary.Metrics.TestsPassed)
	}
	if summary.SummaryText != "\u2713 8 tests passed" {
		t.Errorf("Expected '\u2713 8 tests passed', got '%s'", summary.SummaryText)
	}
}
