package pytest

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

func (p *Parser) Tool() string { return "pytest" }

func (p *Parser) Detect(cmd string, args []string) bool {
	if cmd == "pytest" || strings.HasSuffix(cmd, "pytest") {
		return true
	}
	if cmd == "python" && len(args) >= 2 && args[0] == "-m" && args[1] == "pytest" {
		return true
	}
	return false
}

var resultRegex = regexp.MustCompile(`={2,}\s+(.*?)\s+in\s+([\d\.]+)s\s+={2,}`)

func (p *Parser) Parse(line string) {
	if strings.Contains(line, "FAILURES") || strings.Contains(line, "ERRORS") {
		p.summary.Status = "failure"
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "Found failures/errors section")
	} else if strings.HasPrefix(line, "E   ") {
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
		p.summary.Metrics.Errors++
	}

	match := resultRegex.FindStringSubmatch(line)
	if len(match) > 0 {
		parts := match[1]
		p.summary.Metrics.DurationSeconds, _ = strconv.ParseFloat(match[2], 64)

		for _, part := range strings.Split(parts, ",") {
			part = strings.TrimSpace(part)
			fields := strings.Fields(part)
			if len(fields) > 0 {
				val, _ := strconv.Atoi(fields[0])
				if strings.Contains(part, "passed") {
					p.summary.Metrics.TestsPassed = val
				} else if strings.Contains(part, "failed") {
					p.summary.Metrics.TestsFailed = val
					p.summary.Status = "failure"
				} else if strings.Contains(part, "warning") {
					p.summary.Metrics.Warnings = val
				}
			}
		}
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.TestsFailed > 0 || p.summary.Metrics.Errors > 0 {
		p.summary.Status = "failure"
		p.summary.SummaryText = "\u2717 Pytest failed: " + strconv.Itoa(p.summary.Metrics.TestsFailed) + " failed"
	} else if p.summary.Metrics.TestsPassed > 0 {
		p.summary.SummaryText = "\u2713 Pytest passed: " + strconv.Itoa(p.summary.Metrics.TestsPassed) + " passed"
	} else {
		p.summary.SummaryText = "Pytest execution completed."
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return strings.Contains(line, "test session starts") || strings.Contains(line, "pytest") || (strings.Contains(line, "===") && strings.Contains(line, "passed in"))
}
