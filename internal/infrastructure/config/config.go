package config

import (
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	ServerPort        int
	Timeout           int
	MaxLogLines       int
	FastMode          bool
	FastAnalysisLines int
	AutoFastMode      bool
	AutoFastTrigger   int
	Debug             bool
}

// LoadConfig loads the configuration taking environment variables as highest priority
// falling back to defaults.
func LoadConfig() Config {
	cfg := Config{
		ServerPort:        8080,
		Timeout:           600,
		MaxLogLines:       10000,
		FastMode:          false,
		FastAnalysisLines: 2500,
		AutoFastMode:      true,
		AutoFastTrigger:   6000,
		Debug:             false,
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

	if f := os.Getenv("LOGIQ_FAST_MODE"); f == "true" || f == "1" {
		cfg.FastMode = true
	}

	if l, err := strconv.Atoi(os.Getenv("LOGIQ_FAST_ANALYSIS_LINES")); err == nil {
		cfg.FastAnalysisLines = l
	}

	if f := os.Getenv("LOGIQ_FAST_AUTO_MODE"); f == "false" || f == "0" {
		cfg.AutoFastMode = false
	}

	if l, err := strconv.Atoi(os.Getenv("LOGIQ_FAST_AUTO_TRIGGER_LINES")); err == nil {
		cfg.AutoFastTrigger = l
	}

	if d := os.Getenv("LOGIQ_DEBUG"); d == "true" || d == "1" {
		cfg.Debug = true
	}

	return cfg
}
