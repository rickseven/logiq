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

func TestLooksLikePowerShellCmdlet(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected bool
	}{
		{name: "valid cmdlet", cmd: "Get-ChildItem", expected: true},
		{name: "valid mixed case", cmd: "Set-Location", expected: true},
		{name: "plain executable", cmd: "npm", expected: false},
		{name: "path executable", cmd: "C:/tools/git.exe", expected: false},
		{name: "incomplete cmdlet", cmd: "Get-", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := looksLikePowerShellCmdlet(tc.cmd)
			if got != tc.expected {
				t.Fatalf("looksLikePowerShellCmdlet(%q) = %v, want %v", tc.cmd, got, tc.expected)
			}
		})
	}
}

func TestBuildPowerShellCommand(t *testing.T) {
	got := buildPowerShellCommand("Get-ChildItem", []string{"C:/Program Files", "|", "Select-Object", "Name"})
	want := "Get-ChildItem 'C:/Program Files' | Select-Object Name"
	if got != want {
		t.Fatalf("buildPowerShellCommand() = %q, want %q", got, want)
	}
}
