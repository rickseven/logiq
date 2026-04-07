package detector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadGoldenLines(t *testing.T, name string) []string {
	t.Helper()
	path := filepath.Join("testdata", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading golden log %s: %v", path, err)
	}
	lines := strings.Split(string(b), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func TestDetectFromContentPrefersGradleOnGradleLogs(t *testing.T) {
	d := NewDetector()
	logs := loadGoldenLines(t, "gradle_failure_verbose.log")

	p := d.Detect("analyze", []string{}, logs)
	if p.Tool() != "gradle" {
		t.Fatalf("expected gradle parser, got %q", p.Tool())
	}
}

func TestDetectFromContentStillDetectsJest(t *testing.T) {
	d := NewDetector()
	logs := loadGoldenLines(t, "jest_failure.log")

	p := d.Detect("analyze", []string{}, logs)
	if p.Tool() != "jest" {
		t.Fatalf("expected jest parser, got %q", p.Tool())
	}
}
