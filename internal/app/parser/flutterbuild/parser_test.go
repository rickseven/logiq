package flutterbuild

import (
	"testing"
)

func TestFlutterBuildParser(t *testing.T) {
	parser := NewParser()
	lines := []string{
		"Building without soundness...",
		"Built build/app/outputs/flutter-apk/app-release.apk (14.3MB).",
	}

	for _, line := range lines {
		parser.Parse(line)
	}

	summary := parser.Summary()

	if summary.Status != "success" {
		t.Errorf("Expected success, got %s", summary.Status)
	}
	if summary.Metrics.ArtifactSize != "14.3MB" {
		t.Errorf("Expected 14.3MB artifact size, got '%s'", summary.Metrics.ArtifactSize)
	}
	expected := "\u2713 Flutter build succeeded\nAPK size 14.3MB"
	if summary.SummaryText != expected {
		t.Errorf("Expected '%s', got '%s'", expected, summary.SummaryText)
	}
}
