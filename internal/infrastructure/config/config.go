package config

import (
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	ServerPort  int
	Timeout     int
	MaxLogLines int
	Debug       bool
}

// LoadConfig loads the configuration taking environment variables as highest priority
// falling back to defaults.
func LoadConfig() Config {
	cfg := Config{
		ServerPort:  8080,
		Timeout:     600,
		MaxLogLines: 10000,
		Debug:       false,
	}

	if p, err := strconv.Atoi(os.Getenv("LOGIQ_PORT")); err == nil {
		cfg.ServerPort = p
	}

	if t, err := strconv.Atoi(os.Getenv("LOGIQ_TIMEOUT")); err == nil {
		cfg.Timeout = t
	}

	if m, err := strconv.Atoi(os.Getenv("LOGIQ_MAX_LOG_LINES")); err == nil {
		cfg.MaxLogLines = m
	}

	if d := os.Getenv("LOGIQ_DEBUG"); d == "true" || d == "1" {
		cfg.Debug = true
	}

	return cfg
}
