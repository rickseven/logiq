package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rickseven/logiq/internal/domain"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type StdioServer struct {
	reader *bufio.Reader
	writer io.Writer
}

func NewStdioServer() *StdioServer {
	return &StdioServer{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

func (s *StdioServer) Start() {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			}
			return
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(nil, -32700, "Parse error")
			continue
		}

		s.handleRequest(req)
	}
}

func (s *StdioServer) handleRequest(req JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		s.sendResponse(req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    "logiq",
				"version": domain.GetVersion(),
			},
		})
	case "tools/list":
		s.sendResponse(req.ID, map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "run_command",
					"description": "Execute a terminal command and get an AI-optimized summary of results",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"command": map[string]string{
								"type":        "string",
								"description": "The command line to execute",
							},
						},
						"required": []string{"command"},
					},
				},
				{
					"name":        "analyze_logs",
					"description": "Analyze raw terminal logs and generate an AI-friendly optimized summary",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"logs": map[string]string{
								"type":        "string",
								"description": "The raw terminal output string to compress and analyze",
							},
						},
						"required": []string{"logs"},
					},
				},
				{
					"name":        "explain",
					"description": "Generate a semantic explanation of what a command does",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"command": map[string]string{
								"type":        "string",
								"description": "The command to explain",
							},
						},
						"required": []string{"command"},
					},
				},
				{
					"name":        "doctor",
					"description": "Output system diagnostics and project health metrics",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				{
					"name":        "history_query",
					"description": "Search the command execution history using natural language or keywords",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]string{
								"type":        "string",
								"description": "The search query (e.g. 'build failure', 'memory leak')",
							},
						},
						"required": []string{"query"},
					},
				},
			},
		})
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, -32602, "Invalid params")
			return
		}

		switch params.Name {
		case "run_command":
			var args struct {
				Command string `json:"command"`
			}
			json.Unmarshal(params.Arguments, &args)
			s.runCommandTool(req.ID, args.Command)
		case "analyze_logs":
			var args struct {
				Logs string `json:"logs"`
			}
			json.Unmarshal(params.Arguments, &args)
			s.analyzeLogsTool(req.ID, args.Logs)
		case "explain":
			var args struct {
				Command string `json:"command"`
			}
			json.Unmarshal(params.Arguments, &args)
			s.explainTool(req.ID, args.Command)
		case "doctor":
			s.doctorTool(req.ID)
		case "history_query":
			var args struct {
				Query string `json:"query"`
			}
			json.Unmarshal(params.Arguments, &args)
			s.historyQueryTool(req.ID, args.Query)
		default:
			s.sendError(req.ID, -32601, "Tool not found")
		}
	case "notifications/initialized":
		// Ignore
	default:
		s.sendError(req.ID, -32601, "Method not found")
	}
}

func (s *StdioServer) runCommandTool(id interface{}, cmdLine string) {
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		s.sendToolResult(id, "Error: empty command", true)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	appSvc := getAppService()
	out, err := appSvc.ExecuteCommand(ctx, parts[0], parts[1:], false)

	if err != nil {
		s.sendToolResult(id, fmt.Sprintf("Error: %v", err), true)
		return
	}

	appSvc.RecordTrace(out.ExecutionID, cmdLine, out.Status, out.Summary)

	// Professional Markdown Output
	var sb strings.Builder
	statusIcon := "\u2713"
	if out.Status != "success" {
		statusIcon = "\u2717"
	}

	sb.WriteString(fmt.Sprintf("### Execution Result: %s %s\n\n", statusIcon, out.Status))
	sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", out.Summary))
	sb.WriteString("#### Metrics\n")
	sb.WriteString(fmt.Sprintf("- **Duration:** %.2fs\n", out.Metrics.DurationSeconds))
	if out.Metrics.MaxRAMMB > 0 {
		sb.WriteString(fmt.Sprintf("- **Max RAM:** %.1f MB\n", out.Metrics.MaxRAMMB))
	}
	if out.Metrics.AvgCPUPercent > 0 {
		sb.WriteString(fmt.Sprintf("- **Avg CPU:** %.1f%%\n", out.Metrics.AvgCPUPercent))
	}
	if out.Metrics.TestsPassed > 0 || out.Metrics.TestsFailed > 0 {
		sb.WriteString(fmt.Sprintf("- **Tests:** %d Passed, %d Failed\n", out.Metrics.TestsPassed, out.Metrics.TestsFailed))
	}
	if out.Metrics.Errors > 0 {
		sb.WriteString(fmt.Sprintf("- **Errors Found:** %d\n", out.Metrics.Errors))
	}

	if out.ArtifactPath != "" {
		sb.WriteString(fmt.Sprintf("- **Artifact:** %s\n", out.ArtifactPath))
	}

	if len(out.Suggestions) > 0 {
		sb.WriteString("\n#### Suggested Fixes\n")
		for _, sug := range out.Suggestions {
			sb.WriteString(fmt.Sprintf("- \U0001F4A1 %s\n", sug))
		}
	}

	s.sendToolResult(id, sb.String(), false)
}

func (s *StdioServer) analyzeLogsTool(id interface{}, logs string) {
	appSvc := getAppService()

	// We use the internal app services to process these logs
	// This is where LogIQ shines: Compression + Summarization
	out, _ := appSvc.ExecuteCommand(context.Background(), "analyze", []string{"--raw", logs}, false)

	var sb strings.Builder
	sb.WriteString("### Log Intelligence Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Insight:** %s\n\n", out.Summary))

	if out.Metrics.Errors > 0 || out.Metrics.Warnings > 0 {
		sb.WriteString("#### Detection Stats\n")
		sb.WriteString(fmt.Sprintf("- Errors: %d\n- Warnings: %d\n", out.Metrics.Errors, out.Metrics.Warnings))
	}

	s.sendToolResult(id, sb.String(), false)
}

func (s *StdioServer) explainTool(id interface{}, cmdLine string) {
	appSvc := getAppService()
	parts := strings.Fields(cmdLine)
	res := appSvc.Explain(parts)

	s.sendToolResult(id, fmt.Sprintf("### Command Explanation: `%s`\n\n%s", cmdLine, res.Description), false)
}

func (s *StdioServer) doctorTool(id interface{}) {
	appSvc := getAppService()
	res := appSvc.Diagnose()

	var sb strings.Builder
	sb.WriteString("### LogIQ System Diagnostics\n\n")

	sb.WriteString("| Tool | Status |\n")
	sb.WriteString("| :--- | :--- |\n")
	sb.WriteString(fmt.Sprintf("| **Node.js** | %s |\n", res.Node))
	sb.WriteString(fmt.Sprintf("| **NPM** | %s |\n", res.Npm))
	sb.WriteString(fmt.Sprintf("| **Project** | %s |\n", res.ProjectType))
	if res.Vite != "not found" {
		sb.WriteString(fmt.Sprintf("| **Vite** | %s |\n", res.Vite))
	}

	sb.WriteString("\n#### Yang bisa di-handle LogIQ:\n")
	for i, cap := range res.Capabilities {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, cap))
	}

	sb.WriteString("\n#### Yang perlu diperhatikan (TIDAK bisa / Limitasi):\n")
	for _, lim := range res.Limitations {
		sb.WriteString(fmt.Sprintf("- %s\n", lim))
	}

	s.sendToolResult(id, sb.String(), false)
}

func (s *StdioServer) historyQueryTool(id interface{}, query string) {
	appSvc := getAppService()
	results := appSvc.QueryHistory(query)

	if len(results) == 0 {
		s.sendToolResult(id, "No history entries found matching your query.", false)
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### History Search Results for: `%s`\n\n", query))
	for _, res := range results {
		statusIcon := "\u2713"
		if res.Status != "success" {
			statusIcon = "\u2717"
		}
		sb.WriteString(fmt.Sprintf("- **%s %s**: `%s`\n", statusIcon, res.Status, res.Command))
		if res.Summary != "" {
			sb.WriteString(fmt.Sprintf("  - %s\n", res.Summary))
		}
		if res.ExecutionID != "" {
			sb.WriteString(fmt.Sprintf("  - ID: `%s`\n", res.ExecutionID))
		}
	}

	s.sendToolResult(id, sb.String(), false)
}

func (s *StdioServer) sendToolResult(id interface{}, text string, isError bool) {
	s.sendResponse(id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": text,
			},
		},
		"isError": isError,
	})
}

func (s *StdioServer) sendResponse(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.writeJSON(resp)
}

func (s *StdioServer) sendError(id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	s.writeJSON(resp)
}

func (s *StdioServer) writeJSON(v interface{}) {
	enc := json.NewEncoder(s.writer)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}
