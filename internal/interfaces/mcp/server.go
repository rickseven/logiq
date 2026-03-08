package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rickseven/logiq/internal/app/compress"
	"github.com/rickseven/logiq/internal/app/optimizer"
	"github.com/rickseven/logiq/internal/app/pipeline"
	"github.com/rickseven/logiq/internal/app/summarizer"
	"github.com/rickseven/logiq/internal/infrastructure/observability/metrics"
)

type Server struct {
	port int
}

func NewServer(port int) *Server {
	return &Server{port: port}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Middleware for recovery and charset
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Fprintf(os.Stderr, "Panic in MCP server: %v\n", rec)
				s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server panic"})
			}
		}()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		mux.ServeHTTP(w, r)
	})

	// Existing MCP tools endpoints
	mux.HandleFunc("/tools/run_command", s.handleRunCommand)
	mux.HandleFunc("/tools/run_tests", s.handleRunTests)
	mux.HandleFunc("/tools/build_project", s.handleBuildProject)
	mux.HandleFunc("/tools/analyze_logs", s.handleAnalyzeLogs)

	// New Agent API endpoints
	mux.HandleFunc("/run", s.handleRun)
	mux.HandleFunc("/explain", s.handleExplain)
	mux.HandleFunc("/doctor", s.handleDoctor)
	mux.HandleFunc("/trace", s.handleTrace)
	mux.HandleFunc("/metrics", metrics.HandleMetrics)

	fmt.Fprintf(os.Stderr, "LogIQ MCP server running on :%d\n", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), handler)
}

func (s *Server) writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(v)
}

// POST /tools/run_command
type RunCommandReq struct {
	Command string `json:"command"`
}

func (s *Server) handleRunCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RunCommandReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parts := strings.Fields(req.Command)
	if len(parts) == 0 {
		http.Error(w, "Empty command", http.StatusBadRequest)
		return
	}

	cmd := parts[0]
	args := parts[1:]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	appSvc := getAppService()
	out, err := appSvc.ExecuteCommand(ctx, cmd, args, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"status":  out.Status,
		"summary": out.Summary,
		"metrics": out.Metrics,
	}
	s.writeJSON(w, http.StatusOK, res)
}

// POST /tools/run_tests
type RunTestsReq struct {
	Framework string `json:"framework"`
}

func (s *Server) handleRunTests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RunTestsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var cmd string
	var args []string

	switch strings.ToLower(req.Framework) {
	case "pytest":
		cmd = "pytest"
	case "gotest":
		fallthrough
	case "go":
		cmd = "go"
		args = []string{"test", "./..."}
	case "jest":
		cmd = "npm"
		args = []string{"test"}
	default:
		cmd = req.Framework
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	appSvc := getAppService()
	out, err := appSvc.ExecuteCommand(ctx, cmd, args, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"status":       out.Status,
		"tests_passed": out.Metrics.TestsPassed,
		"tests_failed": out.Metrics.TestsFailed,
	}
	s.writeJSON(w, http.StatusOK, res)
}

// POST /tools/build_project
type BuildProjectReq struct {
	BuildTool string `json:"build_tool"`
}

func (s *Server) handleBuildProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req BuildProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var cmd string
	var args []string

	switch strings.ToLower(req.BuildTool) {
	case "cargo":
		cmd = "cargo"
		args = []string{"build"}
	case "npm":
		cmd = "npm"
		args = []string{"run", "build"}
	case "go":
		cmd = "go"
		args = []string{"build", "./..."}
	default:
		cmd = req.BuildTool
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	appSvc := getAppService()
	out, err := appSvc.ExecuteCommand(ctx, cmd, args, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"status":  out.Status,
		"summary": out.Summary,
		"metrics": out.Metrics,
	}
	s.writeJSON(w, http.StatusOK, res)
}

// POST /tools/analyze_logs
type AnalyzeLogsReq struct {
	Logs string `json:"logs"`
}

func (s *Server) handleAnalyzeLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AnalyzeLogsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lines := strings.Split(req.Logs, "\n")
	logProcessor := pipeline.NewProcessor()

	var rawLogs []string
	for _, rawLine := range lines {
		cleanLine, keep := logProcessor.Process(rawLine)
		if keep {
			rawLogs = append(rawLogs, cleanLine)
		}
	}

	compressedLogs := compress.Compress(rawLogs)
	opt := optimizer.NewOptimizer()
	optimizedLogs := opt.Optimize(compressedLogs)

	summary := summarizer.Summarize(optimizedLogs, 0)

	res := map[string]interface{}{
		"summary":  summary.SummaryText,
		"errors":   summary.Metrics.Errors,
		"warnings": summary.Metrics.Warnings,
	}
	s.writeJSON(w, http.StatusOK, res)
}
