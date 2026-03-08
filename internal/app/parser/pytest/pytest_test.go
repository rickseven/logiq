package pytest

import (
	"strings"
	"testing"
)

func TestPytestParser(t *testing.T) {
	parser := NewParser()

	// Tool Name
	if parser.Tool() != "pytest" {
		t.Errorf("Expected tool to be pytest, got %s", parser.Tool())
	}

	// Detect Test
	if !parser.Detect("pytest", []string{}) {
		t.Errorf("Failed to detect pytest")
	}
	if !parser.Detect("python", []string{"-m", "pytest", "tests/"}) {
		t.Errorf("Failed to detect python -m pytest")
	}

	// Parse Logic
	parser.Parse("==== FAILURES ====")
	parser.Parse("E   AssertionError: assert False")
	// Result summary match
	parser.Parse("==== 2 failed, 15 passed, 1 warnings in 4.34s ====")

	summary := parser.Summary()

	if summary.Status != "failure" {
		t.Errorf("Expected status failure, got %s", summary.Status)
	}
	if summary.Metrics.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", summary.Metrics.Errors)
	}
	if summary.Metrics.TestsFailed != 2 {
		t.Errorf("Expected 2 failed tests, got %d", summary.Metrics.TestsFailed)
	}
	if summary.Metrics.TestsPassed != 15 {
		t.Errorf("Expected 15 passed tests, got %d", summary.Metrics.TestsPassed)
	}
	if summary.Metrics.Warnings != 1 {
		t.Errorf("Expected 1 warning, got %d", summary.Metrics.Warnings)
	}
	if summary.Metrics.DurationSeconds != 4.34 {
		t.Errorf("Expected duration 4.34s, got %v", summary.Metrics.DurationSeconds)
	}

	if !strings.Contains(summary.SummaryText, "Pytest failed: 2 failed") {
		t.Errorf("Unexpected summary text: %q", summary.SummaryText)
	}
}

func TestPytestDetectFromContent(t *testing.T) {
	parser := NewParser()

	if !parser.DetectFromContent("=== test session starts ===") {
		t.Errorf("Failed to detect 'test session starts'")
	}
	if !parser.DetectFromContent("==== 1 failed, 2 passed in 0.12s ====") {
		t.Errorf("Failed to detect summary line")
	}
	if parser.DetectFromContent("no relevant text") {
		t.Errorf("Should not detect irrelevant text")
	}
}
