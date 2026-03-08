package vite

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

func (p *Parser) Tool() string { return "vite" }

func (p *Parser) Detect(cmd string, args []string) bool {
	if cmd == "vite" {
		return true
	}
	if cmd == "npm" && len(args) >= 2 && args[0] == "run" && (args[1] == "build" || args[1] == "dev") {
		return true
	}
	return false
}

func (p *Parser) DetectFromContent(line string) bool {
	lower := strings.ToLower(line)
	return strings.Contains(line, "[vite]") || strings.Contains(lower, "vite v")
}

var (
	modulesTransformedRegex = regexp.MustCompile(`.*?\s*(\d+)\s+modules\s+(?:transformed|compiled)`)
	bundleSizeRegex         = regexp.MustCompile(`dist/.*?(\d+\.\d+\s*[KMG]B)`)
	errorSign               = regexp.MustCompile(`(?i)(?:ERROR|FAILED|Build failed)`)
)

func (p *Parser) Parse(line string) {
	lower := strings.ToLower(line)

	// Detection of errors
	if errorSign.MatchString(line) {
		// Only mark failure if it doesn't look like a success line containing "failed" in a different context
		if !strings.Contains(lower, "0 failed") && !strings.Contains(lower, "operation succeeded") {
			p.summary.Status = "failure"
			p.summary.Metrics.Errors++
			if len(p.summary.ImportantEvents) < 5 {
				p.summary.ImportantEvents = append(p.summary.ImportantEvents, line)
			}
		}
	} else if strings.Contains(lower, "warn") || strings.Contains(lower, "warning") {
		p.summary.Metrics.Warnings++
	}

	// Metrics
	if match := modulesTransformedRegex.FindStringSubmatch(line); match != nil {
		val, _ := strconv.Atoi(match[1])
		p.summary.Metrics.ModulesCompiled = val
	}

	if match := bundleSizeRegex.FindStringSubmatch(line); match != nil {
		p.summary.Metrics.BundleSize = match[1]
		// If we see a bundle size, the build ALMOST CERTAINLY succeeded despite non-fatal errors
		if p.summary.Status == "failure" && p.summary.Metrics.Errors < 2 {
			// This is a debatable heuristic, but often Vite prints errors from a watcher
			// that don't block the final build output.
			// However, if the user explicitly provided a log with ERROR, we should probably keep it.
			// Let's at least mark it as "Succeeded with errors" in Summary.
		}
	}

	if strings.Contains(lower, "build complete") || strings.Contains(lower, "compiled successfully") {
		// Success signal!
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()

	if p.summary.Metrics.BundleSize != "" {
		p.summary.SummaryText = fmt.Sprintf("\u2713 Build completed. Output: %s", p.summary.Metrics.BundleSize)
		if p.summary.Status == "failure" {
			p.summary.SummaryText += " (with non-fatal errors detected)"
		}
	} else if p.summary.Metrics.ModulesCompiled > 0 {
		p.summary.SummaryText = fmt.Sprintf("\u2713 Build succeeded. %d modules compiled", p.summary.Metrics.ModulesCompiled)
	} else if p.summary.Status == "failure" {
		p.summary.SummaryText = "\u2717 Vite build failed"
	} else {
		p.summary.SummaryText = "\u2713 Operation succeeded"
	}
	return p.summary
}
