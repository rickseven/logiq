package playwright

import (
	"strings"
	"testing"
)

func TestPlaywrightParser(t *testing.T) {
	parser := NewParser()

	if parser.Tool() != "playwright" {
		t.Errorf("Expected playful Tool(), got %s", parser.Tool())
	}

	if !parser.Detect("npx playwright test", []string{}) {
		t.Errorf("Should detect playwright command")
	}

	// Parsing Error
	parser.Parse(" Error: Timeout of 30000ms exceeded.")
	// Parsing Summary
	parser.Parse("1 passed (12.3s) 2 failed")

	summary := parser.Summary()

	if summary.Status != "failure" {
		t.Errorf("Expected status failure, got %s", summary.Status)
	}
	if summary.Metrics.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", summary.Metrics.Errors)
	}
	if summary.Metrics.TestsPassed != 1 {
		t.Errorf("Expected 1 passed, got %d", summary.Metrics.TestsPassed)
	}
	if summary.Metrics.TestsFailed != 2 {
		t.Errorf("Expected 2 failed, got %d", summary.Metrics.TestsFailed)
	}
	if !strings.Contains(summary.SummaryText, "Playwright failed: 2 failed") {
		t.Errorf("Unexpected summary text: %q", summary.SummaryText)
	}
}
