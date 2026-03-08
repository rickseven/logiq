package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

// InitLogger initializes the global structured logger
func InitLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	// Always emit structured JSON to stdout/stderr in production environments natively
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// Info logs an informational message
func Info(module, command, status, msg string) {
	if Log == nil {
		InitLogger(false)
	}
	Log.Info(msg,
		slog.String("module", module),
		slog.String("command", command),
		slog.String("status", status),
	)
}

// Error logs an error message
func Error(module, command, errStr string) {
	if Log == nil {
		InitLogger(false)
	}
	Log.Error("An error occurred",
		slog.String("module", module),
		slog.String("command", command),
		slog.String("error", errStr),
	)
}
