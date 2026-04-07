package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rickseven/logiq/internal/app/cmdintel"
	"github.com/rickseven/logiq/internal/app/compress"
	ctxengine "github.com/rickseven/logiq/internal/app/context"
	"github.com/rickseven/logiq/internal/app/debugassist"
	"github.com/rickseven/logiq/internal/app/detector"
	"github.com/rickseven/logiq/internal/app/errorintel"
	"github.com/rickseven/logiq/internal/app/optimizer"
	"github.com/rickseven/logiq/internal/app/pipeline"

	"github.com/rickseven/logiq/internal/domain"
	"github.com/rickseven/logiq/internal/infrastructure/config"
	"github.com/rickseven/logiq/internal/infrastructure/observability/logger"
	"github.com/rickseven/logiq/internal/infrastructure/observability/metrics"
)

// RuntimeEngine defines the infrastructure dependency for running shell commands.
type RuntimeEngine interface {
	Run(ctx context.Context, cmd string, args []string) (<-chan string, <-chan int, *domain.ResourceMetrics, error)
}

// Service orchestrates application logic by coordinating domain behaviors.
// It bridges user interfaces (CLI/MCP) with the execution/analysis pipeline.
type Service struct {
	runner RuntimeEngine
}

type executionTuning struct {
	MaxLogLines       int
	FastMode          bool
	FastAnalysisLines int
	AutoFastMode      bool
	AutoFastTrigger   int
}

type fastModeState struct {
	Triggered     bool
	Kind          string
	TriggerReason string
}

func resolveExecutionTuning() executionTuning {
	cfg := config.LoadConfig()

	maxLogLines := cfg.MaxLogLines
	if maxLogLines <= 0 {
		maxLogLines = 10000
	}

	fastAnalysisLines := cfg.FastAnalysisLines
	if fastAnalysisLines <= 0 {
		fastAnalysisLines = 2500
	}

	if fastAnalysisLines > maxLogLines {
		fastAnalysisLines = maxLogLines
	}

	autoFastTrigger := cfg.AutoFastTrigger
	if autoFastTrigger <= 0 {
		autoFastTrigger = 6000
	}

	if autoFastTrigger > maxLogLines {
		autoFastTrigger = maxLogLines
	}

	return executionTuning{
		MaxLogLines:       maxLogLines,
		FastMode:          cfg.FastMode,
		FastAnalysisLines: fastAnalysisLines,
		AutoFastMode:      cfg.AutoFastMode,
		AutoFastTrigger:   autoFastTrigger,
	}
}

func tailWindow(logs []string, size int) []string {
	if size <= 0 || len(logs) <= size {
		return logs
	}
	return logs[len(logs)-size:]
}

func fastModeMessage(mode string, lines int) string {
	return fmt.Sprintf("[LOGIQ] fast mode active (%s): deep analysis limited to last %d lines", mode, lines)
}

func resolveFastModeState(tuning executionTuning, keptLogLines int) fastModeState {
	manualFastTriggered := tuning.FastMode && keptLogLines > tuning.FastAnalysisLines
	autoFastTriggered := tuning.AutoFastMode && !tuning.FastMode && keptLogLines > tuning.AutoFastTrigger

	state := fastModeState{}
	if manualFastTriggered {
		state.Triggered = true
		state.Kind = "manual"
		state.TriggerReason = "manual_override"
		return state
	}

	if autoFastTriggered {
		state.Triggered = true
		state.Kind = "auto"
		state.TriggerReason = "auto_threshold"
	}

	return state
}

func applyFastModeMetrics(m *domain.Metrics, tuning executionTuning, state fastModeState) {
	m.FastModeActive = state.Triggered
	m.FastModeTriggerLines = tuning.AutoFastTrigger

	if state.Triggered {
		m.FastModeKind = state.Kind
		m.FastAnalysisLines = tuning.FastAnalysisLines
		m.FastModeTriggerReason = state.TriggerReason
	}
}

// NewService provisions the core application orchestrator.
func NewService(runner RuntimeEngine) *Service {
	return &Service{
		runner: runner,
	}
}

// ExecuteCommand drives the main execution lifecycle: running, parsing, compressing, and structuring analytics.
func (s *Service) ExecuteCommand(ctx context.Context, cmd string, args []string, debug bool) (domain.StructuredOutput, error) {
	metrics.IncCommandsExecuted()
	execStartTime := time.Now()

	var rawStream <-chan string
	var exitChan <-chan int
	var resMetrics *domain.ResourceMetrics
	var err error

	executionID := cmdintel.GenerateExecutionID()

	// Virtual command handling: 'analyze --raw <logs>' for MCP/Tools
	if cmd == "analyze" && len(args) >= 2 && args[0] == "--raw" {
		out := make(chan string, 100)
		exit := make(chan int, 1)

		go func() {
			lines := strings.Split(args[1], "\n")
			for _, line := range lines {
				out <- line
			}
			close(out)
			exit <- 0
			close(exit)
		}()
		rawStream = out
		exitChan = exit
		resMetrics = &domain.ResourceMetrics{}
	} else {
		rawStream, exitChan, resMetrics, err = s.runner.Run(ctx, cmd, args)
		if err != nil {
			return domain.StructuredOutput{}, err
		}
	}

	logProcessor := pipeline.NewProcessor()
	var rawLogs []string
	var sampleLogs []string
	tuning := resolveExecutionTuning()
	maxLogLines := tuning.MaxLogLines
	omittedLines := 0

	// Listen to stream with heartbeat
	lastHeartbeat := time.Now()
	for rawLine := range rawStream {
		if debug {
			logger.Log.Debug("[DEBUG RAW]", "line", rawLine)
		}

		if time.Since(lastHeartbeat) > 5*time.Second {
			logger.Log.Info("[LOGIQ HEARTBEAT] still running...", "duration", time.Since(lastHeartbeat).String())
			lastHeartbeat = time.Now()
		}

		cleanLine, keep := logProcessor.Process(rawLine)
		if keep {
			if len(rawLogs) < maxLogLines {
				rawLogs = append(rawLogs, cleanLine)
				if len(sampleLogs) < 10 {
					sampleLogs = append(sampleLogs, cleanLine)
				}
			} else {
				omittedLines++
			}
		}
	}

	if omittedLines > 0 {
		rawLogs = append(rawLogs, fmt.Sprintf("[LOGIQ] omitted %d log lines after LOGIQ_MAX_LOG_LINES limit", omittedLines))
	}

	analysisLogs := rawLogs
	fastState := resolveFastModeState(tuning, len(rawLogs))
	fastModeTriggered := fastState.Triggered
	fastModeKind := fastState.Kind
	if fastModeTriggered {
		analysisLogs = tailWindow(rawLogs, tuning.FastAnalysisLines)
		analysisLogs = append(analysisLogs, fastModeMessage(fastModeKind, tuning.FastAnalysisLines))
	}

	exitCode := <-exitChan
	totalDuration := time.Since(execStartTime)

	// Smart Detection: using both command and log samples
	det := detector.NewDetector()
	parserPlugin := det.Detect(cmd, args, sampleLogs)

	if debug {
		logger.Log.Debug("[DEBUG LOGIQ] Detected parser", "tool", parserPlugin.Tool())
	}

	// 2. Parser sees raw logs for accurate metrics and error detection
	startParser := time.Now()
	for _, rawLine := range analysisLogs {
		parserPlugin.Parse(rawLine)
	}
	metrics.RecordParserExecutionTime(time.Since(startParser))

	// 3. Compress and Optimize for AI consumption and context
	compressedLogs := compress.Compress(analysisLogs)
	opt := optimizer.NewOptimizer()
	optimizedLogs := opt.Optimize(compressedLogs)

	summary := parserPlugin.Summary()
	summary.Metrics.DurationSeconds = totalDuration.Seconds()
	applyFastModeMetrics(&summary.Metrics, tuning, fastState)
	summary.Metrics.MaxRAMMB = resMetrics.MaxRAMMB
	summary.Metrics.AvgCPUPercent = resMetrics.AvgCPUPercent

	artifactLogs := rawLogs
	if fastModeTriggered {
		artifactLogs = analysisLogs
	}
	artifactPath := cmdintel.SaveArtifact(executionID, artifactLogs)

	// Status Logic: exit code based OR parser detected failure
	if exitCode != 0 || summary.Status == "failure" {
		if exitCode != 0 {
			metrics.IncCommandFailures()
			trimmedSummary := strings.TrimSpace(summary.SummaryText)
			if trimmedSummary == "" || !strings.HasPrefix(trimmedSummary, "✗") {
				summary.SummaryText = "✗ Execution failed (exit code " + strconv.Itoa(exitCode) + ")"
			}
		}
		summary.Status = "failure"

		if fastModeTriggered {
			summary.SummaryText += "\n\nFast mode active (" + fastModeKind + "): deep error intelligence skipped. Set LOGIQ_FAST_MODE=0 for full root-cause analysis."
		} else {
			changedFiles := cmdintel.GetChangedFiles()
			intel := errorintel.Analyze(optimizedLogs, changedFiles)
			if intel != nil {
				summary.ErrorIntel = intel
				summary.SummaryText += "\n\nRoot cause:\n" + intel.RootCause
				summary.Suggestions = debugassist.Analyze(intel)
			}
		}
	}

	result := domain.StructuredOutput{
		ExecutionID:     executionID,
		Tool:            parserPlugin.Tool(),
		Command:         cmd,
		Status:          summary.Status,
		Summary:         summary.SummaryText,
		ImportantEvents: summary.ImportantEvents,
		Metrics:         summary.Metrics,
		ErrorIntel:      summary.ErrorIntel,
		Suggestions:     summary.Suggestions,
		ArtifactPath:    artifactPath,
		Timestamp:       time.Now().Format(time.RFC3339),
	}

	comp := ctxengine.NewCompressor(optimizedLogs)
	result.CompressedContext = comp.Compress(&result)

	// --- Token Savings Calculation ---
	var originalBytes int
	for _, line := range rawLogs {
		originalBytes += len(line) + 1 // +1 for newline
	}
	compressedBytes := len(result.CompressedContext)

	result.Metrics.OriginalBytes = originalBytes
	result.Metrics.CompressedBytes = compressedBytes
	if originalBytes > 0 {
		savings := float64(originalBytes-compressedBytes) / float64(originalBytes) * 100.0
		if savings < 0 {
			savings = 0
		}
		result.Metrics.SavingsPercentage = savings
	}

	// --- Tee Mode (Fallback for errors) ---
	// If the command failed, explicitly mark the RawLogPath for AI to retrieve full context
	if result.Status == "failure" {
		result.RawLogPath = artifactPath
	}

	logger.Info("runtime", cmd, result.Status, result.Summary)
	return result, nil
}

// Explain translates commands semantically via predefined heuristics mapped.
func (s *Service) Explain(args []string) domain.ExplainResult {
	return cmdintel.GenerateExplainResult(args)
}

// Diagnose interrogates standard dependency structures globally mapped.
func (s *Service) Diagnose() domain.DoctorResult {
	return cmdintel.GenerateDoctorResult()
}

// History surfaces the traced internal artifacts iteratively.
func (s *Service) History() []domain.TraceEntry {
	entries := cmdintel.GetTraceEntries()
	if entries == nil {
		return []domain.TraceEntry{}
	}
	return entries
}

// QueryHistory searches the history for specific entries.
func (s *Service) QueryHistory(query string) []domain.TraceEntry {
	return cmdintel.QueryHistory(query)
}

// RecordTrace explicitly binds successful executions universally appending state.
func (s *Service) RecordTrace(id, cmd, status, summary string) {
	cmdintel.RecordTrace(id, cmd, status, summary)
}
