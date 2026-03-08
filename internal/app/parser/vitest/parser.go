package vitest

import (
	"fmt"
	"regexp"
	"strconv"
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

func (p *Parser) Tool() string { return "vitest" }

func (p *Parser) Detect(cmd string, args []string) bool {
	if cmd == "vitest" || cmd == "jest" {
		return true
	}
	if cmd == "npm" && len(args) >= 2 && args[0] == "run" && args[1] == "test" {
		return true
	}
	return false
}

var testFilesPassedRegex = regexp.MustCompile(`Test Files\s+(\d+)\s+passed`)
var testsPassedRegex = regexp.MustCompile(`Tests\s+(\d+)\s+passed`)
var testsFailedRegex = regexp.MustCompile(`Tests\s+(\d+)\s+failed`)
var testsPassingRegex = regexp.MustCompile(`(\d+)\s+passing`)
var testsFailingRegex = regexp.MustCompile(`(\d+)\s+failing`)

func (p *Parser) Parse(line string) {
	if strings.Contains(line, "FAIL") || strings.Contains(line, "failed") {
		// Just a heuristic to flag status
		if strings.HasPrefix(line, "FAIL ") || strings.HasPrefix(line, "\u2717 ") {
			p.summary.Status = "failure"
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
		}
	}

	if match := testFilesPassedRegex.FindStringSubmatch(line); match != nil {
		val, _ := strconv.Atoi(match[1])
		p.summary.Metrics.TestFiles = val
	}
	if match := testsPassedRegex.FindStringSubmatch(line); match != nil {
		val, _ := strconv.Atoi(match[1])
		p.summary.Metrics.TestsPassed = val
	}
	if match := testsFailedRegex.FindStringSubmatch(line); match != nil {
		val, _ := strconv.Atoi(match[1])
		p.summary.Metrics.TestsFailed = val
		p.summary.Status = "failure"
	}
	// Fallbacks for jest / mocha
	if match := testsPassingRegex.FindStringSubmatch(line); match != nil {
		val, _ := strconv.Atoi(match[1])
		p.summary.Metrics.TestsPassed = val
	}
	if match := testsFailingRegex.FindStringSubmatch(line); match != nil {
		val, _ := strconv.Atoi(match[1])
		p.summary.Metrics.TestsFailed = val
		p.summary.Status = "failure"
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()

	if p.summary.Metrics.TestsFailed > 0 {
		p.summary.Status = "failure"
		p.summary.SummaryText = fmt.Sprintf("\u2717 %d tests failed", p.summary.Metrics.TestsFailed)
	} else if p.summary.Metrics.TestsPassed > 0 {
		// Ensure status is success if tests passed and no failures recorded
		if p.summary.Status == "failure" && p.summary.Metrics.TestsFailed == 0 {
			// Maybe it was just a transient Fail line that later passed?
			// Normally in vitest FAIL means at least one failed.
		}
		p.summary.SummaryText = fmt.Sprintf("\u2713 %d tests passed", p.summary.Metrics.TestsPassed)
	} else {
		p.summary.SummaryText = "Test execution completed"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return strings.Contains(line, "vitest") || strings.Contains(line, "Test Files") || strings.Contains(line, "PASS ") || strings.Contains(line, "FAIL ")
}
