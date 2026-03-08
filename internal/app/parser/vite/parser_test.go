package vite

import (
	"testing"
)

func TestViteParser(t *testing.T) {
	parser := NewParser()
	lines := []string{
		"vite v5 building for production...",
		"\u2713 142 modules transformed",
		"Build completed in 2.4s",
	}

	for _, line := range lines {
		parser.Parse(line)
	}

	summary := parser.Summary()

	if summary.Status != "success" {
		t.Errorf("Expected success, got %s", summary.Status)
	}
	if summary.Metrics.ModulesCompiled != 142 {
		t.Errorf("Expected 142 modules compiled, got %d", summary.Metrics.ModulesCompiled)
	}
}
