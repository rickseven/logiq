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

func TestRunnerWindowsCmdWithSpacedCommand(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}

	runner := NewRunner()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream, exitChan, _, err := runner.Run(ctx, "cmd", []string{"/c", "echo hello"})
	if err != nil {
		t.Fatalf("failed to run: %v", err)
	}

	found := false
	for line := range stream {
		if strings.Contains(strings.ToLower(line), "hello") {
			found = true
		}
	}

	exitCode := <-exitChan
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !found {
		t.Fatalf("did not find expected output 'hello'")
	}
}

func TestContainsPowerShellSyntax(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		args     []string
		expected bool
	}{
		{
			name:     "Select-String in args",
			cmdName:  "cd",
			args:     []string{`C:\Projects\app;`, "Select-String", "-Pattern", `"^version:"`, "-Path", "pubspec.yaml"},
			expected: true,
		},
		{
			name:     "Get-Content in args",
			cmdName:  "cd",
			args:     []string{`C:\Projects\app;`, "Get-Content", "file.txt"},
			expected: true,
		},
		{
			name:     "$env: variable syntax",
			cmdName:  "echo",
			args:     []string{"$env:PATH"},
			expected: true,
		},
		{
			name:     "ForEach-Object pipeline",
			cmdName:  "dir",
			args:     []string{"|", "ForEach-Object", "{", "$_.Name", "}"},
			expected: true,
		},
		{
			name:     "plain cmd command",
			cmdName:  "dir",
			args:     []string{`C:\Projects`},
			expected: false,
		},
		{
			name:     "findstr is cmd-native",
			cmdName:  "findstr",
			args:     []string{"/R", `"^version:"`, "pubspec.yaml"},
			expected: false,
		},
		{
			name:     "echo is cross-shell",
			cmdName:  "echo",
			args:     []string{"hello", "world"},
			expected: false,
		},
		{
			name:     "flutter is cross-shell",
			cmdName:  "flutter",
			args:     []string{"build", "apk"},
			expected: false,
		},
		{
			name:     "Test-Path in args",
			cmdName:  "cd",
			args:     []string{`C:\app`, "&&", "Test-Path", "build"},
			expected: true,
		},
		{
			name:     "PS comparison operator -eq",
			cmdName:  "if",
			args:     []string{"($x", "-eq", "5)", "{", "echo", "yes", "}"},
			expected: true,
		},
		{
			name:     "Where-Object in pipeline",
			cmdName:  "Get-Process",
			args:     []string{"|", "Where-Object", "{", "$_.CPU", "-gt", "10", "}"},
			expected: true,
		},
		{
			name:     "Invoke-WebRequest (not in old hardcoded list test)",
			cmdName:  "echo",
			args:     []string{";", "Invoke-WebRequest", "https://example.com"},
			expected: true,
		},
		{
			name:     "Uncommon cmdlet Compress-Archive",
			cmdName:  "Compress-Archive",
			args:     []string{"-Path", "src", "-DestinationPath", "out.zip"},
			expected: true,
		},
		{
			name:     "Uncommon cmdlet Expand-Archive",
			cmdName:  "cmd",
			args:     []string{";", "Expand-Archive", "-Path", "file.zip"},
			expected: true,
		},
		{
			name:     "Enable-WindowsOptionalFeature",
			cmdName:  "Enable-WindowsOptionalFeature",
			args:     []string{"-Online", "-FeatureName", "Microsoft-Hyper-V"},
			expected: true,
		},
		{
			name:     "PS type cast [System.IO]::ReadAllText",
			cmdName:  "echo",
			args:     []string{"[System.IO.File]::ReadAllText('file.txt')"},
			expected: true,
		},
		{
			name:     "PS hashtable @{} literal",
			cmdName:  "echo",
			args:     []string{"@{Key='Value'}"},
			expected: true,
		},
		{
			name:     "PS subexpression $()",
			cmdName:  "echo",
			args:     []string{"$(Get-Date)"},
			expected: true,
		},
		{
			name:     "PS $_ pipeline variable",
			cmdName:  "ls",
			args:     []string{"|", "ForEach", "{", "$_.FullName", "}"},
			expected: true,
		},
		{
			name:     "PS -match operator",
			cmdName:  "echo",
			args:     []string{"'hello'", "-match", "'he.*'"},
			expected: true,
		},
		{
			name:     "PS variable ${complex}",
			cmdName:  "echo",
			args:     []string{"${my-var}"},
			expected: true,
		},
		{
			name:     "git is cross-shell",
			cmdName:  "git",
			args:     []string{"status", "--short"},
			expected: false,
		},
		{
			name:     "npm run is cross-shell",
			cmdName:  "npm",
			args:     []string{"run", "build"},
			expected: false,
		},
		{
			name:     "python is cross-shell",
			cmdName:  "python",
			args:     []string{"-m", "pytest"},
			expected: false,
		},
		{
			name:     "dart is cross-shell",
			cmdName:  "dart",
			args:     []string{"analyze"},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := containsPowerShellSyntax(tc.cmdName, tc.args)
			if got != tc.expected {
				t.Errorf("containsPowerShellSyntax(%q, %v) = %v, want %v", tc.cmdName, tc.args, got, tc.expected)
			}
		})
	}
}
