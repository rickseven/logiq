package dotnet

import (
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

func (p *Parser) Tool() string { return "dotnet" }

func (p *Parser) Detect(cmd string, args []string) bool {
	return strings.Contains(cmd, "dotnet") || strings.Contains(cmd, "msbuild") || strings.HasSuffix(cmd, "nuget")
}

var errorRegex = regexp.MustCompile(`(.*?)\((\d+),(\d+)\): error (CS\d+): (.*) \[(.*)\]`)
var summaryRegex = regexp.MustCompile(`(\d+) Error\(s\).*?(\d+) Warning\(s\)`)

func (p *Parser) Parse(line string) {
	if strings.Contains(line, ": error CS") {
		p.summary.Status = "failure"
		p.summary.Metrics.Errors++
		match := errorRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+match[1]+":"+match[2]+" - "+match[5])
		} else {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+strings.TrimSpace(line))
		}
	} else if strings.Contains(line, ": warning CS") {
		p.summary.Metrics.Warnings++
	}

	if match := summaryRegex.FindStringSubmatch(line); match != nil {
		errs, _ := strconv.Atoi(match[1])
		warns, _ := strconv.Atoi(match[2])
		p.summary.Metrics.Errors = errs
		p.summary.Metrics.Warnings = warns
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()
	if p.summary.Metrics.Errors > 0 {
		p.summary.SummaryText = "\u2717 .NET build failed with " + strconv.Itoa(p.summary.Metrics.Errors) + " errors"
		p.summary.Status = "failure"
	} else {
		p.summary.SummaryText = "\u2713 .NET operation successful"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
