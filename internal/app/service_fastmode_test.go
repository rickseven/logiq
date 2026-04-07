package app

import (
	"testing"

	"github.com/rickseven/logiq/internal/domain"
)

func TestResolveFastModeState(t *testing.T) {
	tests := []struct {
		name         string
		tuning       executionTuning
		keptLogLines int
		want         fastModeState
	}{
		{
			name: "manual override triggers when above analysis window",
			tuning: executionTuning{
				FastMode:          true,
				FastAnalysisLines: 100,
				AutoFastMode:      true,
				AutoFastTrigger:   80,
			},
			keptLogLines: 101,
			want: fastModeState{
				Triggered:     true,
				Kind:          "manual",
				TriggerReason: "manual_override",
			},
		},
		{
			name: "auto threshold triggers when manual is off",
			tuning: executionTuning{
				FastMode:          false,
				FastAnalysisLines: 100,
				AutoFastMode:      true,
				AutoFastTrigger:   80,
			},
			keptLogLines: 81,
			want: fastModeState{
				Triggered:     true,
				Kind:          "auto",
				TriggerReason: "auto_threshold",
			},
		},
		{
			name: "manual takes precedence over auto",
			tuning: executionTuning{
				FastMode:          true,
				FastAnalysisLines: 10,
				AutoFastMode:      true,
				AutoFastTrigger:   1,
			},
			keptLogLines: 50,
			want: fastModeState{
				Triggered:     true,
				Kind:          "manual",
				TriggerReason: "manual_override",
			},
		},
		{
			name: "no trigger when below thresholds",
			tuning: executionTuning{
				FastMode:          false,
				FastAnalysisLines: 100,
				AutoFastMode:      true,
				AutoFastTrigger:   80,
			},
			keptLogLines: 80,
			want:         fastModeState{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveFastModeState(tc.tuning, tc.keptLogLines)
			if got != tc.want {
				t.Fatalf("resolveFastModeState() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestApplyFastModeMetrics(t *testing.T) {
	tuning := executionTuning{
		FastAnalysisLines: 1200,
		AutoFastTrigger:   3000,
	}

	t.Run("sets fast mode fields when active", func(t *testing.T) {
		metrics := domain.Metrics{}
		state := fastModeState{
			Triggered:     true,
			Kind:          "auto",
			TriggerReason: "auto_threshold",
		}

		applyFastModeMetrics(&metrics, tuning, state)

		if !metrics.FastModeActive {
			t.Fatalf("FastModeActive = false, want true")
		}
		if metrics.FastModeKind != "auto" {
			t.Fatalf("FastModeKind = %q, want %q", metrics.FastModeKind, "auto")
		}
		if metrics.FastAnalysisLines != 1200 {
			t.Fatalf("FastAnalysisLines = %d, want %d", metrics.FastAnalysisLines, 1200)
		}
		if metrics.FastModeTriggerLines != 3000 {
			t.Fatalf("FastModeTriggerLines = %d, want %d", metrics.FastModeTriggerLines, 3000)
		}
		if metrics.FastModeTriggerReason != "auto_threshold" {
			t.Fatalf("FastModeTriggerReason = %q, want %q", metrics.FastModeTriggerReason, "auto_threshold")
		}
	})

	t.Run("keeps trigger lines but omits fast-specific fields when inactive", func(t *testing.T) {
		metrics := domain.Metrics{}
		state := fastModeState{}

		applyFastModeMetrics(&metrics, tuning, state)

		if metrics.FastModeActive {
			t.Fatalf("FastModeActive = true, want false")
		}
		if metrics.FastModeKind != "" {
			t.Fatalf("FastModeKind = %q, want empty", metrics.FastModeKind)
		}
		if metrics.FastAnalysisLines != 0 {
			t.Fatalf("FastAnalysisLines = %d, want 0", metrics.FastAnalysisLines)
		}
		if metrics.FastModeTriggerReason != "" {
			t.Fatalf("FastModeTriggerReason = %q, want empty", metrics.FastModeTriggerReason)
		}
		if metrics.FastModeTriggerLines != 3000 {
			t.Fatalf("FastModeTriggerLines = %d, want %d", metrics.FastModeTriggerLines, 3000)
		}
	})
}
