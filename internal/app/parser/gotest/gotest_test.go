package gotest

import (
	"testing"
)

func TestGoTestParser(t *testing.T) {
	parser := NewParser()

	lines := []string{
		"ok  \tlogiq/internal/parser/pytest\t0.123s",
		"--- FAIL: TestGoTestParser (0.00s)",
		"FAIL\tlogiq/internal/parser/gotest\t0.200s",
	}

	for _, line := range lines {
		parser.Parse(line)
	}

	summary := parser.Summary()

	if summary.Status != "failure" {
		t.Errorf("expected status 'failure', got '%s'", summary.Status)
	}

	if summary.Metrics.TestsFailed != 1 {
		t.Errorf("expected 1 failed test, got %d", summary.Metrics.TestsFailed)
	}
}
