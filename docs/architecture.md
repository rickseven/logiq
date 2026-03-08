# LogIQ Architecture

LogIQ has been refactored strictly into a **Clean Architecture** to sustainably support future AI agent integrations, CLI extensions, and standalone API services (MCP). By explicitly separating logic into functional boundaries, modules no longer rigidly block system updates.

## Design Philosophy

The project enforces robust `Dependency Inversion`, enforcing control flows strictly as:
`Interfaces Layer → Application Layer → Infrastructure Layer → Domain Layer`

Infrastructure components implement behaviors (like caching or file execution), but only Application logic dictates _how_ domain entities utilize those outputs.

## Directory Structure & Responsibilities

### 1. `internal/domain/` (The Core Entities)

Holds the absolute lowest-level fundamental models and strict interfaces for LogIQ logic. These models reflect purely theoretical objects uncoupled from networking or file systems.

- Data structures are discretely separated into pure behaviorless contexts like `stream.go`, `metrics.go`, `error.go`, `summary.go`, `output.go`, `cmdintel.go`, and `plugin.go`.
- Contains definitions such as `domain.StructuredOutput`, `domain.Metrics`, `domain.Plugin`, `domain.Parser`, and `domain.ErrorIntel`.
- Completely zero outbound dependencies.

### 2. `internal/app/` (Application Layer)

Houses LogIQ's exclusive algorithmic systems and runtime coordination. It resolves domain transformations across sub-packages.

- `app.Service`: The centralized orchestrator mapping raw execution into LogIQ's core lifecycle (Execution → Parse → Compress → Optimize → Summarize & Diff).
- Explicit standalone domain subcomponents: `cmdintel/`, `compress/`, `context/`, `debugassist/`, `detector/`, `errorintel/`, `optimizer/`, `pipeline/`, and `summarizer/`.

### 3. `internal/interfaces/` (Delivery Mechanisms)

The boundary interfacing exactly with the outside consumer.

- `cli/output/`: Decodes agent JSON models or human-readable stdout behaviors.
- `mcp/`: Implements the Model Context Protocol exposing secure HTTP JSON endpoints to connect upstream LLM applications without knowing internal configurations.

### 4. `internal/infrastructure/` (Supporting Implementations)

Exclusively handles complex side-effecting code, caching, metric emissions, and external physical runtime interactions.

- `artifacts/`: Writes file-level summary snapshots.
- `cache/`: Speeds up deterministic computations (e.g. `logiq doctor` or `explain`).
- `config/`: Secure unmarshaling of environmental boundaries.
- `observability/`: Handles `logger` structure bindings and `metrics` mutexes exclusively.
- `plugin/`: Handles external compiled implementations dynamic loading limits seamlessly mapping directly to App integrations.
- `runtime/`: Governs `exec.Command` streaming bindings directly.

### 5. `plugins/` (Extensions)

Standalone integration definitions injecting foreign parser tools dynamically at compile time seamlessly (such as `flutter` and `vue`).

## Orchestration Flow

1. An agent queries `logiq run npm run build`.
2. `cmd/logiq/main.go` parses arguments uniquely triggering the `app.Service`.
3. `app.Service` dispatches process streaming execution to `infrastructure/runtime`.
4. The output is streamed down continuously mutating inside `app/...` utilities (detector, optimizer, context).
5. The unified `domain.StructuredOutput` is built directly returning backwards safely.
6. The CLI interface (or MCP listener) converts it correctly rendering safely back out explicitly bypassing core state risks.
