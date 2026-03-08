package cypress

import (
	"fmt"
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

func (p *Parser) Tool() string { return "cypress" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return strings.Contains(cmd, "cypress")
}

func (p *Parser) Parse(line string) {
	if strings.Contains(line, " (failed)") || strings.Contains(line, " ✖ ") {
		p.summary.Status = "failure"
		p.summary.Metrics.TestsFailed++
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+strings.TrimSpace(line))
	} else if strings.Contains(line, " (passed)") || strings.Contains(line, " ✔ ") {
		p.summary.Metrics.TestsPassed++
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.TestsFailed > 0 {
		p.summary.Status = "failure"
		p.summary.SummaryText = fmt.Sprintf("\u2717 Cypress failed: %d failing", p.summary.Metrics.TestsFailed)
	} else if p.summary.Metrics.TestsPassed > 0 {
		p.summary.SummaryText = fmt.Sprintf("\u2713 Cypress passed: %d passing", p.summary.Metrics.TestsPassed)
	} else {
		p.summary.SummaryText = "Cypress execution completed"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return strings.Contains(line, "cypress") || strings.Contains(line, "Run Finished") || strings.Contains(line, "Tests Results:")
}
