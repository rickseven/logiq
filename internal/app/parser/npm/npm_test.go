package npm

import (
	"strings"
	"testing"
)

func TestNPMParser(t *testing.T) {
	parser := NewParser()

	// Initial State
	if parser.Tool() != "npm" {
		t.Errorf("Expected tool to be npm, got %s", parser.Tool())
	}

	// Detect Test
	if !parser.Detect("npm", []string{"run", "build"}) {
		t.Errorf("Failed to detect simple npm command")
	}

	// Parse errors
	parser.Parse("npm ERR! missing script: test")
	parser.Parse("npm error A generic error")
	parser.Parse("npm WARN skipping package")
	parser.Parse("npm warn another warning")
	parser.Parse("1 failing")
	parser.Parse("1 passing")
	parser.Parse("1 passing")
	parser.Parse("1 passing")

	summary := parser.Summary()

	// Assertions
	if summary.Status != "failure" {
		t.Errorf("Expected status failure, got %s", summary.Status)
	}
	if summary.Metrics.Errors != 2 {
		t.Errorf("Expected 2 errors, got %d", summary.Metrics.Errors)
	}
	if summary.Metrics.Warnings != 2 {
		t.Errorf("Expected 2 warnings, got %d", summary.Metrics.Warnings)
	}
	if summary.Metrics.TestsFailed != 1 {
		t.Errorf("Expected 1 failed test, got %d", summary.Metrics.TestsFailed)
	}
	if summary.Metrics.TestsPassed != 3 {
		t.Errorf("Expected 3 passed tests, got %d", summary.Metrics.TestsPassed)
	}
	if !strings.Contains(summary.SummaryText, "Missing script error") {
		t.Errorf("Expected Missing script error in summary text, got %q", summary.SummaryText)
	}
}

func TestNPMDetectFromContent(t *testing.T) {
	parser := NewParser()

	if !parser.DetectFromContent("npm ERR! some error") {
		t.Errorf("Should detect 'npm ERR!' ")
	}

	if !parser.DetectFromContent("npm error missing script") {
		t.Errorf("Should detect 'npm error' ")
	}

	if parser.DetectFromContent("no relevant info here") {
		t.Errorf("Should not detect irrelevant lines")
	}
}
