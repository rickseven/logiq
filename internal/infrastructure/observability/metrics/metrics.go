package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	CommandsExecutedTotal int64
	CommandFailuresTotal  int64
	ParserExecutionTime   int64
	mu                    sync.RWMutex
)

// IncCommandsExecuted increments the total commands executed
func IncCommandsExecuted() {
	mu.Lock()
	defer mu.Unlock()
	CommandsExecutedTotal++
}

// IncCommandFailures increments the total commands failed
func IncCommandFailures() {
	mu.Lock()
	defer mu.Unlock()
	CommandFailuresTotal++
}

// RecordParserExecutionTime records parser times
func RecordParserExecutionTime(d time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	ParserExecutionTime += int64(d.Milliseconds())
}

// HandleMetrics exposes metrics to an HTTP endpoint
func HandleMetrics(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	data := map[string]interface{}{
		"commands_executed_total": CommandsExecutedTotal,
		"command_failures_total":  CommandFailuresTotal,
		"parser_execution_time":   ParserExecutionTime,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
