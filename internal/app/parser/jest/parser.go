package jest

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

func (p *Parser) Tool() string { return "jest" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return cmd == "jest" || strings.HasSuffix(cmd, "jest")
}

var testSummaryRegex = regexp.MustCompile(`Tests:\s+(\d+) failed,\s+(\d+) passed,\s+(\d+) total`)

func (p *Parser) Parse(line string) {
	if strings.Contains(line, " FAIL ") {
		p.summary.Status = "failure"
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+strings.TrimSpace(line))
	}

	match := testSummaryRegex.FindStringSubmatch(line)
	if len(match) > 0 {
		fVal, _ := strconv.Atoi(match[1])
		pVal, _ := strconv.Atoi(match[2])
		p.summary.Metrics.TestsFailed = fVal
		p.summary.Metrics.TestsPassed = pVal
		if fVal > 0 {
			p.summary.Status = "failure"
		}
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.TestsFailed > 0 {
		p.summary.Status = "failure"
		p.summary.SummaryText = "\u2717 Jest failed: " + strconv.Itoa(p.summary.Metrics.TestsFailed) + " failed"
	} else if p.summary.Metrics.TestsPassed > 0 {
		p.summary.SummaryText = "\u2713 Jest passed: " + strconv.Itoa(p.summary.Metrics.TestsPassed) + " passed"
	} else {
		p.summary.SummaryText = "Jest execution completed."
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return strings.Contains(line, "jest") || strings.Contains(line, "FAIL") || strings.Contains(line, "PASS") || strings.Contains(line, "Test Suites:")
}
