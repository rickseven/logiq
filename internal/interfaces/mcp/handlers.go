package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"regexp"
	stdruntime "runtime"
	"strings"
	"unicode"

	"github.com/rickseven/logiq/internal/app"
	"github.com/rickseven/logiq/internal/domain"
	"github.com/rickseven/logiq/internal/infrastructure/runtime"
)

var shellControlPattern = regexp.MustCompile(`[\n\r;|><]|&&|\|\|`)
var psControlPattern = regexp.MustCompile(`[;\n\r]`)
var psCmdletPattern = regexp.MustCompile(`(?i)\b[A-Za-z][A-Za-z0-9]*-[A-Za-z][A-Za-z0-9]*\b`)
var psSyntaxPattern = regexp.MustCompile(`(?i)\$env:|\$\{|\$\(|@\{|@\(|\$_\.|\s-(?:eq|ne|gt|lt|ge|le|like|notlike|match|notmatch|contains|in|notin|and|or|not)\s`)

func shouldRunAsRawShell(command string) bool {
	return shellControlPattern.MatchString(command)
}

func looksLikePowerShellCommand(command string) bool {
	if psControlPattern.MatchString(command) {
		return true
	}
	if psCmdletPattern.MatchString(command) {
		return true
	}
	if psSyntaxPattern.MatchString(command) {
		return true
	}
	return false
}

func resolvePowerShellExe() string {
	if _, err := exec.LookPath("pwsh"); err == nil {
		return "pwsh"
	}
	return "powershell"
}

func splitCommandLine(command string) []string {
	runes := []rune(command)
	parts := make([]string, 0, 8)
	var current strings.Builder
	inQuote := rune(0)

	flush := func() {
		if current.Len() == 0 {
			return
		}
		parts = append(parts, current.String())
		current.Reset()
	}

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		if inQuote != 0 {
			if ch == inQuote {
				inQuote = 0
				continue
			}
			if ch == '\\' && inQuote == '"' && i+1 < len(runes) {
				next := runes[i+1]
				if next == '"' || next == '\\' {
					current.WriteRune(next)
					i++
					continue
				}
			}
			current.WriteRune(ch)
			continue
		}

		if ch == '"' || ch == '\'' {
			inQuote = ch
			continue
		}

		if unicode.IsSpace(ch) {
			flush()
			continue
		}

		current.WriteRune(ch)
	}

	flush()
	return parts
}

// parseCommandForExecution preserves complex shell syntax on Windows by routing
// raw command text to the matching shell. For simple commands it keeps argv split
// to preserve parser detection quality in the app pipeline.
func parseCommandForExecution(command string) (string, []string) {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return "", nil
	}

	if stdruntime.GOOS == "windows" && shouldRunAsRawShell(trimmed) {
		if looksLikePowerShellCommand(trimmed) {
			psExe := resolvePowerShellExe()
			return psExe, []string{"-NoProfile", "-NonInteractive", "-Command", trimmed}
		}
		return "cmd", []string{"/C", trimmed}
	}

	parts := splitCommandLine(trimmed)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

func getAppService() *app.Service {
	// Reusable binding abstraction explicitly matching CLI bindings
	cmdRunner := runtime.NewRunner()
	return app.NewService(cmdRunner)
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RunCommandReq // reusing struct from server.go
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd, args := parseCommandForExecution(req.Command)
	if cmd == "" {
		http.Error(w, "Empty command", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpToolTimeout())
	defer cancel()

	defer func() {
		if rec := recover(); rec != nil {
			s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server panic"})
		}
	}()

	appSvc := getAppService()
	out, err := appSvc.ExecuteCommand(ctx, cmd, args, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appSvc.RecordTrace(out.ExecutionID, req.Command, out.Status, out.Summary)

	res := map[string]interface{}{
		"status":  out.Status,
		"summary": out.Summary,
		"metrics": out.Metrics,
	}
	s.writeJSON(w, http.StatusOK, res)
}

type ExplainReq struct {
	Command string `json:"command"`
}

func (s *Server) handleExplain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ExplainReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parts := strings.Fields(req.Command)
	if len(parts) == 0 {
		http.Error(w, "Empty command", http.StatusBadRequest)
		return
	}

	appSvc := getAppService()
	result := appSvc.Explain(parts)
	result.Description = strings.Split(result.Description, "\n")[0]

	s.writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appSvc := getAppService()
	result := appSvc.Diagnose()

	s.writeJSON(w, http.StatusOK, map[string]string{
		"node":         result.Node,
		"npm":          result.Npm,
		"vite":         result.Vite,
		"flutter":      result.Flutter,
		"project_type": result.ProjectType,
	})
}

func (s *Server) handleTrace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appSvc := getAppService()
	entries := appSvc.History()
	if entries == nil {
		entries = []domain.TraceEntry{}
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"commands": entries,
	})
}
