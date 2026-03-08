package cmdintel

import (
	"os"
	"testing"
)

func TestExplainResult(t *testing.T) {
	res := GenerateExplainResult([]string{"npm", "run", "build"})

	if res.Tool != "vite" {
		t.Errorf("Expected vite tool mapping for npm run build, got %s", res.Tool)
	}
	if res.Type != "build" {
		t.Errorf("Expected build type, got %s", res.Type)
	}

	resFlutter := GenerateExplainResult([]string{"flutter", "test"})
	if resFlutter.Tool != "flutter" {
		t.Errorf("Expected flutter, got %s", resFlutter.Tool)
	}
}

func TestDoctorResultMapping(t *testing.T) {
	// Temporarily create a fake configuration to test Vue project detection
	f, _ := os.Create("vite.config.js")
	defer func() {
		f.Close()
		os.Remove("vite.config.js")
	}()

	res := GenerateDoctorResult()
	if res.ProjectType != "Vue (Vite)" {
		t.Errorf("Expected Vue project type due to vite.config.js, got %s", res.ProjectType)
	}
}

func TestTraceRecording(t *testing.T) {
	// Swap out the actual history file for testing avoiding conflicts
	backup := getHistoryPath()
	os.Remove(backup)

	RecordTrace("test-id", "cmd testing", "success", "executed securely")

	historyPath := getHistoryPath()
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		t.Fatalf("Expected history trace file to be written, but not found")
	}

	entries := GetTraceEntries()
	if len(entries) == 0 {
		t.Fatalf("Expected at least one trace entry")
	}

	os.Remove(historyPath)
}
