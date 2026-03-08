package linter

import (
	"strings"

	"github.com/rickseven/logiq/internal/domain"
)

type Parser struct {
	summary domain.Summary
	tool    string
}

func NewParser() *Parser {
	return &Parser{
		summary: domain.Summary{
			Status:          "success",
			ImportantEvents: make([]string, 0),
		},
	}
}

func (p *Parser) Tool() string { return p.tool }

func (p *Parser) Detect(cmd string, args []string) bool {
	tools := []string{"prettier", "stylelint"}
	for _, t := range tools {
		if strings.Contains(cmd, t) {
			p.tool = t
			return true
		}
	}
	return false
}

func (p *Parser) Parse(line string) {
	lower := strings.ToLower(line)
	if strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
		p.summary.Status = "failure"
		p.summary.Metrics.Errors++
		if len(p.summary.ImportantEvents) < 5 {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u26A0 "+strings.TrimSpace(line))
		}
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.Errors > 0 {
		p.summary.SummaryText = "\u2717 " + p.tool + " found issues"
	} else {
		p.summary.SummaryText = "\u2713 " + p.tool + " passed"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
