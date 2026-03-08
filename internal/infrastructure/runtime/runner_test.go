package runtime

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunner(t *testing.T) {
	// A cross-platform echo command simulation
	runner := NewRunner()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var cmdName string
	var args []string

	if runtime.GOOS == "windows" {
		cmdName = "cmd"
		args = []string{"/c", "echo", "hello"}
	} else {
		cmdName = "echo"
		args = []string{"hello"}
	}

	stream, exitChan, _, err := runner.Run(ctx, cmdName, args)
	if err != nil {
		t.Fatalf("failed to run: %v", err)
	}

	found := false
	for line := range stream {
		t.Logf("line: %q", line)
		if strings.Contains(line, "hello") {
			found = true
		}
	}

	<-exitChan

	if !found {
		t.Errorf("did not find expected output from command")
	}
}
func TestSanitizeUTF8(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello \xff world", "hello  world"}, // invalid byte removed
		{"✓ check", "✓ check"},
	}

	for _, tc := range tests {
		got := sanitizeUTF8(tc.input)
		if got != tc.expected {
			t.Errorf("sanitizeUTF8(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
