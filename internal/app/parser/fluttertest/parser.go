package fluttertest

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

func (p *Parser) Tool() string { return "flutter_test" }

func (p *Parser) Detect(cmd string, args []string) bool {
	if cmd == "flutter" && len(args) > 0 && args[0] == "test" {
		return true
	}
	return false
}

var allPassedRegex = regexp.MustCompile(`\+(\d+):\s*All tests passed!`)
var someFailedRegex = regexp.MustCompile(`\+(\d+)\s*-(\d+):\s*Some tests failed`)

func (p *Parser) Parse(line string) {
	if strings.Contains(line, "EXCEPTION: ") || strings.Contains(line, "Exception: ") {
		p.summary.Status = "failure"
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
	}

	if match := allPassedRegex.FindStringSubmatch(line); match != nil {
		val, _ := strconv.Atoi(match[1])
		p.summary.Metrics.TestsPassed = val
	} else if match := someFailedRegex.FindStringSubmatch(line); match != nil {
		pVal, _ := strconv.Atoi(match[1])
		fVal, _ := strconv.Atoi(match[2])
		p.summary.Metrics.TestsPassed = pVal
		p.summary.Metrics.TestsFailed = fVal
		p.summary.Status = "failure"
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()

	if p.summary.Metrics.TestsFailed > 0 {
		p.summary.Status = "failure"
		p.summary.SummaryText = fmt.Sprintf("\u2717 %d tests failed", p.summary.Metrics.TestsFailed)
	} else if p.summary.Metrics.TestsPassed > 0 {
		p.summary.SummaryText = fmt.Sprintf("\u2713 %d tests passed", p.summary.Metrics.TestsPassed)
	} else {
		p.summary.SummaryText = "Flutter test completed"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return strings.Contains(line, "All tests passed!") || strings.Contains(line, "Some tests failed") || (strings.Contains(line, "Shell:") && strings.Contains(line, "test"))
}
