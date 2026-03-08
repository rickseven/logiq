package golangcilint

import (
	"strings"

	"github.com/rickseven/logiq/internal/domain"
)

type Parser struct {
	summary domain.Summary
}

func NewParser() *Parser {
	return &Parser{
		summary: domain.Summary{
			Status:          "success",
			ImportantEvents: make([]string, 0),
		},
	}
}

func (p *Parser) Tool() string { return "golangci-lint" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return cmd == "golangci-lint" || strings.HasSuffix(cmd, "golangci-lint")
}

func (p *Parser) Parse(line string) {
	if strings.Contains(line, ":") && (strings.Contains(line, "errcheck") || strings.Contains(line, "unused") || strings.Contains(line, "staticcheck")) {
		p.summary.Metrics.Errors++
		p.summary.Status = "failure"
		if len(p.summary.ImportantEvents) < 5 {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u26A0 "+strings.TrimSpace(line))
		}
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.Errors > 0 {
		p.summary.SummaryText = "\u2717 golangci-lint found issues"
	} else {
		p.summary.SummaryText = "\u2713 golangci-lint passed"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
