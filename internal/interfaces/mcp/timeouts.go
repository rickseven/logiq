package mcp

import (
	"os"
	"strconv"
	"time"
)

func timeoutFromEnv(defaultValue time.Duration) time.Duration {
	raw := os.Getenv("LOGIQ_MCP_TIMEOUT_SECONDS")
	if raw == "" {
		return defaultValue
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return defaultValue
	}

	return time.Duration(seconds) * time.Second
}

func stdioToolTimeout() time.Duration {
	return timeoutFromEnv(10 * time.Minute)
}

func httpToolTimeout() time.Duration {
	return timeoutFromEnv(5 * time.Minute)
}
