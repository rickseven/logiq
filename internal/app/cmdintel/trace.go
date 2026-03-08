package cmdintel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rickseven/logiq/internal/domain"
)

func getHistoryPath() string {
	return filepath.Join(".logiq", "history.jsonl")
}

// RecordTrace silently tracks the executed command locally
func RecordTrace(id, cmd, status, summary string) {
	os.MkdirAll(".logiq", 0755)
	f, err := os.OpenFile(getHistoryPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	entry := domain.TraceEntry{ExecutionID: id, Command: cmd, Status: status, Summary: summary}
	bytes, _ := json.Marshal(entry)
	f.Write(bytes)
	f.WriteString("\n")
}

// SaveArtifact stores the raw logs for a command execution
func SaveArtifact(id string, logs []string) string {
	dir := filepath.Join(".logiq", "artifacts")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, fmt.Sprintf("%s.log", id))

	f, err := os.Create(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	for _, line := range logs {
		f.WriteString(line + "\n")
	}
	return path
}

// GetTraceEntries returns the list of recent traces
func GetTraceEntries() []domain.TraceEntry {
	f, err := os.Open(getHistoryPath())
	if err != nil {
		return nil
	}
	defer f.Close()

	var entries []domain.TraceEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry domain.TraceEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			entries = append(entries, entry)
		}
	}
	return entries
}
