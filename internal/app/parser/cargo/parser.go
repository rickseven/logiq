package cargo

import (
	"github.com/rickseven/logiq/internal/domain"
	"regexp"
	"strconv"
	"strings"
	"time"
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

func (p *Parser) Tool() string { return "cargo" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return strings.HasSuffix(cmd, "cargo")
}

var resultRegex = regexp.MustCompile(`test result: (ok|FAILED)\. (\d+) passed; (\d+) failed; (\d+) ignored; (\d+) measured; (\d+) filtered out; finished in ([\d\.]+)s`)

func (p *Parser) Parse(line string) {
	if strings.Contains(line, "error:") || strings.Contains(line, "error[") {
		p.summary.Status = "failure"
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
		p.summary.Metrics.Errors++
	} else if strings.Contains(line, "warning:") {
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
		p.summary.Metrics.Warnings++
	}

	if strings.HasPrefix(line, "test result:") {
		match := resultRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			if match[1] == "FAILED" {
				p.summary.Status = "failure"
			} else {
				p.summary.Status = "success"
			}
			p.summary.Metrics.TestsPassed, _ = strconv.Atoi(match[2])
			p.summary.Metrics.TestsFailed, _ = strconv.Atoi(match[3])
			p.summary.Metrics.DurationSeconds, _ = strconv.ParseFloat(match[7], 64)
		}
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.DurationSeconds == 0 {
		p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()
	}
	p.summary.SummaryText = "Cargo command finished"
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
