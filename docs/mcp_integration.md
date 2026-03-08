# LogIQ MCP Integration

LogIQ can be integrated into your AI agents as a native MCP (Model Context Protocol) tool using the local HTTP Server module.

## Getting Started

1. Start the LogIQ MCP server on your preferred port (default 8080):

```bash
go run cmd/logiq-mcp/main.go --port 8080
```

2. Point your AI agent's tool execution framework to `http://localhost:8080` (or the configured port).

## Available Tools (API Endpoints)

All endpoints accept and return JSON. The server ensures safe concurrency by instantiating new runners internally upon each request.

### 1. `run_command`

Executes an arbitrary shell-like command, analyzes the logs organically via LogIQ, and returns formatted metrics.

- **URL:** `POST /tools/run_command`
- **Request Body:**
  ```json
  {
    "command": "pytest"
  }
  ```
- **Response:**
  ```json
  {
    "status": "success",
    "summary": "4 tests passed",
    "metrics": {
      "tests_passed": 4,
      "tests_failed": 0,
      "warnings": 0,
      "errors": 0,
      "duration_seconds": 1.2
    }
  }
  ```

### 2. `run_tests`

Abstracts specific test framework execution (e.g., `pytest`, `gotest`, `jest`) to find patterns effortlessly.

- **URL:** `POST /tools/run_tests`
- **Request Body:**
  ```json
  {
    "framework": "pytest"
  }
  ```
- **Response:**
  ```json
  {
    "status": "success",
    "tests_passed": 12,
    "tests_failed": 0
  }
  ```

### 3. `build_project`

Automatically configures and runs standard build systems depending on the provided build tool definition (e.g., `cargo`, `npm`, `go`).

- **URL:** `POST /tools/build_project`
- **Request Body:**
  ```json
  {
    "build_tool": "cargo"
  }
  ```
- **Response:**
  ```json
  {
    "status": "success",
    "summary": "Build succeeded"
  }
  ```

### 4. `analyze_logs`

Analyzes raw logs strictly strings with the Semantic Log Summarizer pipeline, allowing LogIQ features passively.

- **URL:** `POST /tools/analyze_logs`
- **Request Body:**
  ```json
  {
    "logs": "Running tests...\n...\nERROR: compilation failed\n"
  }
  ```
- **Response:**
  ```json
  {
    "summary": "❌ Build failed",
    "errors": 1,
    "warnings": 0
  }
  ```
