package fluttertest

import (
	"testing"
)

func TestFlutterTestParser(t *testing.T) {
	parser := NewParser()
	lines := []string{
		"00:01 +12: All tests passed!",
	}

	for _, line := range lines {
		parser.Parse(line)
	}

	summary := parser.Summary()

	if summary.Status != "success" {
		t.Errorf("Expected success, got %s", summary.Status)
	}
	if summary.Metrics.TestsPassed != 12 {
		t.Errorf("Expected 12 tests passed, got %d", summary.Metrics.TestsPassed)
	}
	if summary.SummaryText != "\u2713 12 tests passed" {
		t.Errorf("Expected '\u2713 12 tests passed', got '%s'", summary.SummaryText)
	}
}
