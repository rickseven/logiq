package eslint

import (
	"testing"
)

func TestEslintParser(t *testing.T) {
	parser := NewParser()
	lines := []string{
		"12:5 warning unused variable",
		"18:3 error missing semicolon",
	}

	for _, line := range lines {
		parser.Parse(line)
	}

	summary := parser.Summary()

	if summary.Status != "failure" {
		t.Errorf("Expected failure, got %s", summary.Status)
	}
	if summary.Metrics.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", summary.Metrics.Errors)
	}
	if summary.Metrics.Warnings != 1 {
		t.Errorf("Expected 1 warning, got %d", summary.Metrics.Warnings)
	}
	expected := "\u2717 1 error\n\u26A0 1 warning"
	if summary.SummaryText != expected {
		t.Errorf("Expected '%s', got '%s'", expected, summary.SummaryText)
	}
}
