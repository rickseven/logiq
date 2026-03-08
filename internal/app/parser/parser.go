package parser

import "github.com/rickseven/logiq/internal/domain"

// Parser defines the plugin architecture for log parsing
type Parser interface {
	// Tool returns the name of the parser tool
	Tool() string

	// Detect returns whether this parser can handle the given command
	Detect(cmd string, args []string) bool

	// Parse processes a single line of log output
	Parse(line string)

	// Summary returns the extracted semantic meaning
	Summary() domain.Summary
}
