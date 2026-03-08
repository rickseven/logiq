package typescript

import (
	"regexp"
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

func (p *Parser) Tool() string { return "typescript" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return cmd == "tsc" || cmd == "vue-tsc" || strings.HasSuffix(cmd, "tsc")
}

var errorRegex = regexp.MustCompile(`(.*?)\((\d+),(\d+)\): error (TS\d+): (.*)`)

func (p *Parser) Parse(line string) {
	if strings.Contains(line, ": error TS") {
		p.summary.Status = "failure"
		p.summary.Metrics.Errors++
		match := errorRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+match[1]+":"+match[2]+" - "+match[5])
		} else {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+line)
		}
	} else if strings.Contains(line, "Found") && strings.Contains(line, "error") {
		p.summary.Status = "failure"
	}
}

func (p *Parser) Summary() domain.Summary {
	if p.summary.Metrics.Errors > 0 {
		p.summary.SummaryText = "\u2717 Type checking failed: " + strings.Join(p.summary.ImportantEvents, " | ")
		if p.summary.SummaryText == "" {
			p.summary.SummaryText = "Type checking errors found."
		}
	} else {
		p.summary.SummaryText = "\u2713 Type checking passed"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
