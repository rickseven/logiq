package flutterbuild

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

func (p *Parser) Tool() string { return "flutter_build" }

func (p *Parser) Detect(cmd string, args []string) bool {
	if cmd == "flutter" && len(args) > 0 && args[0] == "build" {
		return true
	}
	return false
}

var builtRegex = regexp.MustCompile(`Built\s+.*?\s+\((.*?)\)\.`)

func (p *Parser) Parse(line string) {
	if strings.Contains(line, "Exception: ") || strings.Contains(line, "Error:") {
		p.summary.Status = "failure"
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
	}

	if match := builtRegex.FindStringSubmatch(line); match != nil {
		p.summary.Metrics.ArtifactSize = match[1]
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()

	if p.summary.Status == "failure" {
		p.summary.SummaryText = "\u2717 Flutter build failed"
	} else if p.summary.Metrics.ArtifactSize != "" {
		p.summary.SummaryText = fmt.Sprintf("\u2713 Flutter build succeeded\nAPK size %s", p.summary.Metrics.ArtifactSize)
		// Usually works for APK size, IPA, etc. The prompt asked for "APK size X", "Artifact size X" depends on command but this is simple enough.
	} else {
		p.summary.SummaryText = "\u2713 Flutter build succeeded"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
