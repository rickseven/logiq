package pipeline

import (
	"regexp"
	"strings"
)

type LogProcessor interface {
	Process(line string) (string, bool)
}

type DefaultProcessor struct {
	lastLine string
}

func NewProcessor() LogProcessor {
	return &DefaultProcessor{}
}

var ansiRegex = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
var timestampRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}[\.,]\d{3}Z?\s*`)

func (p *DefaultProcessor) Process(line string) (string, bool) {
	// Strip ANSI
	line = ansiRegex.ReplaceAllString(line, "")

	// Strip common timestamp prefixes
	line = timestampRegex.ReplaceAllString(line, "")

	// Trim spaces
	line = strings.TrimSpace(line)

	// Skip empty lines
	if line == "" {
		return "", false
	}

	// Simple noise reduction: avoid consecutive duplicates
	if line == p.lastLine {
		return "", false
	}
	p.lastLine = line

	if strings.HasPrefix(line, "Downloading ") || strings.HasPrefix(line, "Progress: ") {
		return "", false
	}

	return line, true
}
