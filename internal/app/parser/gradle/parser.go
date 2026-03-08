package gradle

import (
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

func (p *Parser) Tool() string { return "gradle" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return strings.Contains(cmd, "gradle") || strings.HasSuffix(cmd, "gradlew")
}

var taskRegex = regexp.MustCompile(`^> Task :(.*)`)

func (p *Parser) Parse(line string) {
	if strings.HasPrefix(line, "FAILURE: Build failed") || strings.Contains(line, "FAILED") {
		p.summary.Status = "failure"
	}

	if match := taskRegex.FindStringSubmatch(line); match != nil {
		// Log important tasks like compile, build, assemble
		task := match[1]
		if strings.Contains(task, "compile") || strings.Contains(task, "assemble") || strings.Contains(task, "bundle") {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2699 "+task)
		}
	}

	if strings.Contains(line, "error:") || strings.Contains(line, "e: ") {
		p.summary.Metrics.Errors++
		p.summary.Status = "failure"
		if len(p.summary.ImportantEvents) < 10 {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+strings.TrimSpace(line))
		}
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()
	if p.summary.Status == "failure" {
		p.summary.SummaryText = "\u2717 Gradle build failed"
	} else {
		p.summary.SummaryText = "\u2713 Gradle build succeeded"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
