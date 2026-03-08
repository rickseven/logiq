package flutterrun

import (
	"strings"
	"time"

	"github.com/rickseven/logiq/internal/domain"
)

type Parser struct {
	summary domain.Summary
	start   time.Time
}

func NewParser() *Parser {
	return &Parser{
		summary: domain.Summary{
			Status:          "success",
			ImportantEvents: make([]string, 0),
		},
		start: time.Now(),
	}
}

func (p *Parser) Tool() string { return "flutter_run" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return cmd == "flutter" && len(args) > 0 && args[0] == "run"
}

func (p *Parser) Parse(line string) {
	if strings.Contains(line, "Running with unset MSYS") || strings.Contains(line, "Launching") {
		return
	}
	if strings.Contains(line, "Error") || strings.Contains(line, "Exception") {
		p.summary.Status = "failure"
		p.summary.Metrics.Errors++
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+strings.TrimSpace(line))
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()
	if p.summary.Status == "failure" {
		p.summary.SummaryText = "\u2717 Flutter run failed"
	} else {
		p.summary.SummaryText = "\u2713 Flutter app is running"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
