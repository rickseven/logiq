package app

import (
	"context"
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
			rawLogs = append(rawLogs, cleanLine)
			if len(sampleLogs) < 10 {
				sampleLogs = append(sampleLogs, cleanLine)
			}
		}
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
	for _, rawLine := range rawLogs {
		parserPlugin.Parse(rawLine)
	}
	metrics.RecordParserExecutionTime(time.Since(startParser))

	// 3. Compress and Optimize for AI consumption and context
	compressedLogs := compress.Compress(rawLogs)
	opt := optimizer.NewOptimizer()
	optimizedLogs := opt.Optimize(compressedLogs)

	summary := parserPlugin.Summary()
	summary.Metrics.DurationSeconds = totalDuration.Seconds()
	summary.Metrics.MaxRAMMB = resMetrics.MaxRAMMB
	summary.Metrics.AvgCPUPercent = resMetrics.AvgCPUPercent

	artifactPath := cmdintel.SaveArtifact(executionID, rawLogs)

	// Status Logic: exit code based OR parser detected failure
	if exitCode != 0 || summary.Status == "failure" {
		if exitCode != 0 {
			metrics.IncCommandFailures()
		}
		summary.Status = "failure"

		changedFiles := cmdintel.GetChangedFiles()
		intel := errorintel.Analyze(optimizedLogs, changedFiles)
		if intel != nil {
			summary.ErrorIntel = intel
			summary.SummaryText += "\n\nRoot cause:\n" + intel.RootCause
			summary.Suggestions = debugassist.Analyze(intel)
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
