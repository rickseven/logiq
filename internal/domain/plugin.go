package domain

import "regexp"

// Parser defines the plugin architecture for log parsing
type Parser interface {
	Tool() string
	Detect(cmd string, args []string) bool
	DetectFromContent(line string) bool
	Parse(line string)
	Summary() Summary
}

// SuggestionRule defines a pattern to match against an error root cause
type SuggestionRule struct {
	Pattern    *regexp.Regexp
	ErrorType  string
	Suggestion string
}

// Plugin interface allows external modules to extend LogIQ without modifying core code.
type Plugin interface {
	Name() string
	Parsers() []Parser
	DebugRules() []SuggestionRule
}
