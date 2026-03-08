package output

import (
	"encoding/json"
	"fmt"
	"github.com/rickseven/logiq/internal/domain"
)

// PrintExplain handles CLI output for generic semantic explanation flows
func PrintExplain(mode string, result domain.ExplainResult) {
	if mode == "json" || mode == "agent" {
		bytes, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(bytes))
		return
	}
	fmt.Printf("%s\n", result.Description)
}

// PrintDoctor handles CLI output for diagnostic outputs
func PrintDoctor(mode string, result domain.DoctorResult) {
	if mode == "json" || mode == "agent" {
		bytes, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(bytes))
		return
	}

	fmt.Println("LogIQ Doctor Diagnostics:")
	fmt.Printf("- Project Type: %s\n", result.ProjectType)
	fmt.Printf("- Node.js:      %s\n", result.Node)
	fmt.Printf("- Npm:          %s\n", result.Npm)
	fmt.Printf("- Vite:         %s\n", result.Vite)
	fmt.Printf("- Flutter:      %s\n", result.Flutter)
	fmt.Printf("- Dart:         %s\n", result.Dart)
	fmt.Printf("- Git:          %s\n", result.Git)
}

// PrintTrace handles CLI traced cache buffers
func PrintTrace(mode string, entries []domain.TraceEntry) {
	if mode == "json" || mode == "agent" {
		bytes, _ := json.MarshalIndent(entries, "", "  ")
		fmt.Println(string(bytes))
		return
	}

	if len(entries) == 0 {
		fmt.Println("No recent traces found locally.")
		return
	}

	fmt.Println("LogIQ Local Traces:")
	for idx, entry := range entries {
		fmt.Printf("[%d] Command: '%s' | Status: %s | Summary: %s\n", idx+1, entry.Command, entry.Status, entry.Summary)
	}
}
