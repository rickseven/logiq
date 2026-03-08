package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rickseven/logiq/internal/app"
	"github.com/rickseven/logiq/internal/domain"
	"github.com/rickseven/logiq/internal/infrastructure/runtime"
)

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

	parts := strings.Fields(req.Command)
	if len(parts) == 0 {
		http.Error(w, "Empty command", http.StatusBadRequest)
		return
	}

	cmd := parts[0]
	args := parts[1:]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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
