package gotest

import (
	"fmt"
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

func (p *Parser) Tool() string { return "go test" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return cmd == "go" && len(args) > 0 && args[0] == "test"
}

func (p *Parser) Parse(line string) {
	if strings.Contains(line, "--- FAIL:") {
		p.summary.Metrics.TestsFailed++
		p.summary.Status = "failure"
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
	} else if strings.Contains(line, "--- PASS:") {
		p.summary.Metrics.TestsPassed++
	} else if strings.HasPrefix(line, "FAIL\t") {
		p.summary.Status = "failure"
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "Test suite failed: "+strings.TrimSpace(line))
	} else if strings.HasPrefix(line, "ok\t") || strings.HasPrefix(line, "PASS") {
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "Test suite passed: "+strings.TrimSpace(line))
	} else if strings.Contains(line, "build failed") || strings.Contains(line, "build errors") {
		p.summary.Status = "failure"
		p.summary.Metrics.Errors++
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "Build error: "+line)
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.TestsFailed > 0 || p.summary.Metrics.Errors > 0 {
		p.summary.Status = "failure"
	}
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()

	if p.summary.Status == "failure" {
		p.summary.SummaryText = fmt.Sprintf("\u2717 Go Tests Failed. Passed: %d, Failed: %d",
			p.summary.Metrics.TestsPassed, p.summary.Metrics.TestsFailed)
	} else {
		p.summary.SummaryText = fmt.Sprintf("\u2713 Go Tests Passed. Passed: %d", p.summary.Metrics.TestsPassed)
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return strings.HasPrefix(line, "--- PASS:") || strings.HasPrefix(line, "--- FAIL:") || strings.HasPrefix(line, "ok\t") || strings.HasPrefix(line, "FAIL\t")
}
