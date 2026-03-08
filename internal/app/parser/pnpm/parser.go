package pnpm

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

func (p *Parser) Tool() string { return "pnpm" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return strings.HasSuffix(cmd, "pnpm")
}

func (p *Parser) Parse(line string) {
	if strings.Contains(line, "ERR_PNPM_") || strings.Contains(line, "ERROR") {
		p.summary.Status = "failure"
		p.summary.Metrics.Errors++
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+strings.TrimSpace(line))
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()
	if p.summary.Status == "failure" {
		p.summary.SummaryText = "\u2717 pnpm command failed"
	} else {
		p.summary.SummaryText = "\u2713 pnpm command successful"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
