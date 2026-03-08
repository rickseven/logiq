package npm

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

func (p *Parser) Tool() string { return "npm" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return cmd == "npm" || strings.HasSuffix(cmd, "npm")
}

func (p *Parser) Parse(line string) {
	// Support both classic "npm ERR!" and modern "npm error"
	lower := strings.ToLower(line)
	if strings.Contains(line, "ERR!") || strings.Contains(lower, "npm error") {
		p.summary.Status = "failure"
		p.summary.Metrics.Errors++
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)

		if strings.Contains(lower, "missing script") {
			p.summary.SummaryText = "\u2717 Missing script error"
		}
	} else if strings.Contains(line, "WARN") || strings.Contains(lower, "npm warn") {
		p.summary.Metrics.Warnings++
	}

	if strings.Contains(line, "failing") {
		p.summary.Status = "failure"
		p.summary.Metrics.TestsFailed++
	} else if strings.Contains(line, "passing") {
		p.summary.Metrics.TestsPassed++
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()

	if p.summary.SummaryText == "" {
		if p.summary.Status == "failure" {
			p.summary.SummaryText = "\u2717 NPM execution failed"
		} else {
			p.summary.SummaryText = "\u2713 NPM execution completed"
		}
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	lower := strings.ToLower(line)
	return strings.Contains(lower, "npm error") || strings.Contains(line, "npm ERR!")
}
