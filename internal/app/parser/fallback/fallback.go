package fallback

import (
	"time"

	"github.com/rickseven/logiq/internal/app/compress"
	"github.com/rickseven/logiq/internal/app/summarizer"
	"github.com/rickseven/logiq/internal/domain"
)

type Parser struct {
	logs  []string
	start time.Time
}

func NewParser() *Parser {
	return &Parser{
		logs:  make([]string, 0),
		start: time.Now(),
	}
}

func (p *Parser) Tool() string { return "fallback" }

func (p *Parser) Detect(cmd string, args []string) bool {
	// Fallback should NOT match by default.
	// The Detector will use it as a last resort.
	return false
}

func (p *Parser) Parse(line string) {
	p.logs = append(p.logs, line)
}

func (p *Parser) Summary() domain.Summary {
	compressed := compress.Compress(p.logs)
	duration := time.Since(p.start)
	return summarizer.Summarize(compressed, duration)
}

func (p *Parser) DetectFromContent(line string) bool {
	// Always false, we don't want to match it from content unless as last resort.
	return false
}
