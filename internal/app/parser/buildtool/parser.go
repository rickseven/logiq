package buildtool

import (
	"strings"
	"time"

	"github.com/rickseven/logiq/internal/domain"
)

type Parser struct {
	summary domain.Summary
	tool    string
	start   time.Time
}

func NewParser() *Parser {
	return &Parser{
		summary: domain.Summary{
			Status:          "success",
			ImportantEvents: make([]string, 0),
		},
		start: time.Now(),
		tool:  "buildtool",
	}
}

func (p *Parser) Tool() string { return p.tool }

func (p *Parser) Detect(cmd string, args []string) bool {
	tools := []string{"webpack", "rollup", "esbuild", "next", "nuxt", "turbopack"}
	for _, t := range tools {
		if strings.Contains(cmd, t) {
			p.tool = t
			return true
		}
	}
	return false
}

func (p *Parser) Parse(line string) {
	lower := strings.ToLower(line)
	if strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
		p.summary.Status = "failure"
		p.summary.Metrics.Errors++
		if len(p.summary.ImportantEvents) < 5 {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+strings.TrimSpace(line))
		}
	} else if strings.Contains(lower, "compiled successfully") || strings.Contains(lower, "build successful") {
		p.summary.Status = "success"
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()
	if p.summary.Status == "failure" {
		p.summary.SummaryText = "\u2717 " + p.tool + " build failed"
	} else {
		p.summary.SummaryText = "\u2713 " + p.tool + " build successful"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	lower := strings.ToLower(line)
	keywords := []string{"webpack", "rollup", "esbuild", "next.js", "nuxt", "compiled successfully"}
	for _, k := range keywords {
		if strings.Contains(lower, k) {
			p.tool = k
			return true
		}
	}
	return false
}
