package eslint

import (
	"fmt"
	"regexp"
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

func (p *Parser) Tool() string { return "eslint" }

func (p *Parser) Detect(cmd string, args []string) bool {
	if cmd == "eslint" {
		return true
	}
	if cmd == "npm" && len(args) >= 2 && args[0] == "run" && strings.Contains(args[1], "lint") {
		return true
	}
	return false
}

var (
	eslintFileRegex = regexp.MustCompile(`\s*(error|warning)\s+`)
	zeroErrorRegex  = regexp.MustCompile(`0\s+errors?`)
)

func (p *Parser) Parse(line string) {
	// Heuristic: actual ESLint error lines usually have " error " (with spaces)
	// and are NOT "0 errors"
	if (strings.Contains(line, " error ") || strings.HasPrefix(strings.TrimSpace(line), "error ")) && !zeroErrorRegex.MatchString(line) {
		p.summary.Metrics.Errors++
		p.summary.Status = "failure"
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
	} else if strings.Contains(line, " warning ") || strings.HasPrefix(strings.TrimSpace(line), "warning ") {
		p.summary.Metrics.Warnings++
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()

	if p.summary.Metrics.Errors > 0 || p.summary.Metrics.Warnings > 0 {
		errStr := ""
		if p.summary.Metrics.Errors > 0 {
			errStr += fmt.Sprintf("\u2717 %d error", p.summary.Metrics.Errors)
			if p.summary.Metrics.Errors > 1 {
				errStr += "s"
			}
		}
		warnStr := ""
		if p.summary.Metrics.Warnings > 0 {
			if errStr != "" {
				warnStr += "\n"
			}
			warnStr += fmt.Sprintf("\u26A0 %d warning", p.summary.Metrics.Warnings)
			if p.summary.Metrics.Warnings > 1 {
				warnStr += "s"
			}
		}
		p.summary.SummaryText = errStr + warnStr
	} else {
		p.summary.SummaryText = "\u2713 No lint errors"
	}

	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return strings.Contains(line, "eslint") || (strings.Contains(line, "error") && strings.Contains(line, "warning") && strings.Contains(line, ":"))
}
