package dart

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

func (p *Parser) Tool() string { return "dart" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return cmd == "dart" && len(args) > 0 && args[0] == "analyze"
}

func (p *Parser) Parse(line string) {
	if strings.HasPrefix(line, "info ") || strings.HasPrefix(line, "error ") || strings.HasPrefix(line, "warning ") {
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
		if strings.HasPrefix(line, "error ") {
			p.summary.Status = "failure"
			p.summary.Metrics.Errors++
		} else if strings.HasPrefix(line, "warning ") {
			p.summary.Metrics.Warnings++
		}
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.Errors > 0 {
		p.summary.SummaryText = "\u2717 Dart analyze failed with " + strings.Repeat("error", 1)
	} else if p.summary.Metrics.Warnings > 0 {
		p.summary.SummaryText = "\u26A0 Dart analyze found warnings"
	} else {
		p.summary.SummaryText = "\u2713 Dart analyze passed"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
