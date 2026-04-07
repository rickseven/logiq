package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/rickseven/logiq/internal/domain"
)

type stubRunner struct {
	lines    []string
	exitCode int
	metrics  *domain.ResourceMetrics
}

func (s *stubRunner) Run(ctx context.Context, cmd string, args []string) (<-chan string, <-chan int, *domain.ResourceMetrics, error) {
	out := make(chan string, len(s.lines))
	exit := make(chan int, 1)

	go func() {
		for _, line := range s.lines {
			select {
			case <-ctx.Done():
				close(out)
				exit <- 1
				close(exit)
				return
			case out <- line:
			}
		}
		close(out)
		exit <- s.exitCode
		close(exit)
	}()

	metrics := s.metrics
	if metrics == nil {
		metrics = &domain.ResourceMetrics{}
	}

	return out, exit, metrics, nil
}

func generateUniqueLogs(lines int) string {
	var b strings.Builder
	for i := 0; i < lines; i++ {
		b.WriteString(fmt.Sprintf("line-%d\n", i))
	}
	return b.String()
}

func withTempWorkingDir(t *testing.T) {
	t.Helper()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to read current working directory: %v", err)
	}

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to switch working directory: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})
}

func TestExecuteCommandFastModeIntegration(t *testing.T) {
	withTempWorkingDir(t)

	tests := []struct {
		name              string
		fastMode          string
		autoFastMode      string
		autoFastTrigger   string
		fastAnalysisLines string
		maxLogLines       string
		logLines          int
		wantActive        bool
		wantKind          string
		wantReason        string
		wantTriggerLines  int
		wantAnalysisLines int
	}{
		{
			name:              "auto fast mode end to end",
			fastMode:          "0",
			autoFastMode:      "1",
			autoFastTrigger:   "3000",
			fastAnalysisLines: "1200",
			maxLogLines:       "10000",
			logLines:          4000,
			wantActive:        true,
			wantKind:          "auto",
			wantReason:        "auto_threshold",
			wantTriggerLines:  3000,
			wantAnalysisLines: 1200,
		},
		{
			name:              "manual fast mode end to end",
			fastMode:          "1",
			autoFastMode:      "1",
			autoFastTrigger:   "3000",
			fastAnalysisLines: "1200",
			maxLogLines:       "10000",
			logLines:          2000,
			wantActive:        true,
			wantKind:          "manual",
			wantReason:        "manual_override",
			wantTriggerLines:  3000,
			wantAnalysisLines: 1200,
		},
		{
			name:              "fast mode disabled end to end",
			fastMode:          "0",
			autoFastMode:      "0",
			autoFastTrigger:   "3000",
			fastAnalysisLines: "1200",
			maxLogLines:       "10000",
			logLines:          4000,
			wantActive:        false,
			wantKind:          "",
			wantReason:        "",
			wantTriggerLines:  3000,
			wantAnalysisLines: 0,
		},
	}

	svc := NewService(nil)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("LOGIQ_FAST_MODE", tc.fastMode)
			t.Setenv("LOGIQ_FAST_AUTO_MODE", tc.autoFastMode)
			t.Setenv("LOGIQ_FAST_AUTO_TRIGGER_LINES", tc.autoFastTrigger)
			t.Setenv("LOGIQ_FAST_ANALYSIS_LINES", tc.fastAnalysisLines)
			t.Setenv("LOGIQ_MAX_LOG_LINES", tc.maxLogLines)

			logs := generateUniqueLogs(tc.logLines)
			out, err := svc.ExecuteCommand(context.Background(), "analyze", []string{"--raw", logs}, false)
			if err != nil {
				t.Fatalf("ExecuteCommand returned error: %v", err)
			}

			if out.Metrics.FastModeActive != tc.wantActive {
				t.Fatalf("FastModeActive = %v, want %v", out.Metrics.FastModeActive, tc.wantActive)
			}
			if out.Metrics.FastModeKind != tc.wantKind {
				t.Fatalf("FastModeKind = %q, want %q", out.Metrics.FastModeKind, tc.wantKind)
			}
			if out.Metrics.FastModeTriggerReason != tc.wantReason {
				t.Fatalf("FastModeTriggerReason = %q, want %q", out.Metrics.FastModeTriggerReason, tc.wantReason)
			}
			if out.Metrics.FastModeTriggerLines != tc.wantTriggerLines {
				t.Fatalf("FastModeTriggerLines = %d, want %d", out.Metrics.FastModeTriggerLines, tc.wantTriggerLines)
			}
			if out.Metrics.FastAnalysisLines != tc.wantAnalysisLines {
				t.Fatalf("FastAnalysisLines = %d, want %d", out.Metrics.FastAnalysisLines, tc.wantAnalysisLines)
			}
		})
	}
}

func TestExecuteCommandFailureSummaryIncludesFastModeMessage(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv("LOGIQ_FAST_MODE", "0")
	t.Setenv("LOGIQ_FAST_AUTO_MODE", "1")
	t.Setenv("LOGIQ_FAST_AUTO_TRIGGER_LINES", "300")
	t.Setenv("LOGIQ_FAST_ANALYSIS_LINES", "100")
	t.Setenv("LOGIQ_MAX_LOG_LINES", "1000")

	raw := generateUniqueLogs(600)
	lines := strings.Split(strings.TrimSuffix(raw, "\n"), "\n")
	svc := NewService(&stubRunner{lines: lines, exitCode: 1})

	out, err := svc.ExecuteCommand(context.Background(), "cmd", []string{"/c", "echo"}, false)
	if err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}

	if out.Status != "failure" {
		t.Fatalf("Status = %q, want %q", out.Status, "failure")
	}
	if !out.Metrics.FastModeActive {
		t.Fatalf("FastModeActive = false, want true")
	}
	if out.Metrics.FastModeKind != "auto" {
		t.Fatalf("FastModeKind = %q, want %q", out.Metrics.FastModeKind, "auto")
	}
	if !strings.Contains(out.Summary, "Fast mode active (auto): deep error intelligence skipped") {
		t.Fatalf("Summary does not contain fast mode failure hint, got: %q", out.Summary)
	}
}

func TestExecuteCommandFailureSummaryIncludesManualFastModeMessage(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv("LOGIQ_FAST_MODE", "1")
	t.Setenv("LOGIQ_FAST_AUTO_MODE", "1")
	t.Setenv("LOGIQ_FAST_AUTO_TRIGGER_LINES", "300")
	t.Setenv("LOGIQ_FAST_ANALYSIS_LINES", "100")
	t.Setenv("LOGIQ_MAX_LOG_LINES", "1000")

	raw := generateUniqueLogs(220)
	lines := strings.Split(strings.TrimSuffix(raw, "\n"), "\n")
	svc := NewService(&stubRunner{lines: lines, exitCode: 1})

	out, err := svc.ExecuteCommand(context.Background(), "cmd", []string{"/c", "echo"}, false)
	if err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}

	if out.Status != "failure" {
		t.Fatalf("Status = %q, want %q", out.Status, "failure")
	}
	if !out.Metrics.FastModeActive {
		t.Fatalf("FastModeActive = false, want true")
	}
	if out.Metrics.FastModeKind != "manual" {
		t.Fatalf("FastModeKind = %q, want %q", out.Metrics.FastModeKind, "manual")
	}
	if out.Metrics.FastModeTriggerReason != "manual_override" {
		t.Fatalf("FastModeTriggerReason = %q, want %q", out.Metrics.FastModeTriggerReason, "manual_override")
	}
	if !strings.Contains(out.Summary, "Fast mode active (manual): deep error intelligence skipped") {
		t.Fatalf("Summary does not contain manual fast mode failure hint, got: %q", out.Summary)
	}
}
