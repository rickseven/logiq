package playwright

import (
	"regexp"
	"strconv"
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

func (p *Parser) Tool() string { return "playwright" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return strings.Contains(cmd, "playwright")
}

var pwSummaryRegex = regexp.MustCompile(`(\d+) passed.*\(.*\)\s+(\d+) failed`)

func (p *Parser) Parse(line string) {
	if strings.Contains(line, " Error: ") {
		p.summary.Status = "failure"
		p.summary.Metrics.Errors++
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+strings.TrimSpace(line))
	}

	match := pwSummaryRegex.FindStringSubmatch(line)
	if len(match) > 0 {
		p.summary.Metrics.TestsPassed, _ = strconv.Atoi(match[1])
		p.summary.Metrics.TestsFailed, _ = strconv.Atoi(match[2])
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.TestsFailed > 0 {
		p.summary.Status = "failure"
		p.summary.SummaryText = "\u2717 Playwright failed: " + strconv.Itoa(p.summary.Metrics.TestsFailed) + " failed"
	} else if p.summary.Metrics.TestsPassed > 0 {
		p.summary.SummaryText = "\u2713 Playwright passed: " + strconv.Itoa(p.summary.Metrics.TestsPassed) + " passed"
	} else {
		p.summary.SummaryText = "Playwright execution completed."
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
