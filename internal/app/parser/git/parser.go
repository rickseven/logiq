package git

import (
	"regexp"
	"strings"

	"github.com/rickseven/logiq/internal/domain"
)

type Parser struct {
	summary domain.Summary
	command string
}

func NewParser() *Parser {
	return &Parser{
		summary: domain.Summary{
			Status:          "success",
			ImportantEvents: make([]string, 0),
		},
	}
}

func (p *Parser) Tool() string { return "git" }

func (p *Parser) Detect(cmd string, args []string) bool {
	if cmd == "git" || strings.HasSuffix(cmd, "git") || strings.HasSuffix(cmd, "git.exe") {
		if len(args) > 0 {
			p.command = args[0]
		}
		return true
	}
	return false
}

var diffStatRegex = regexp.MustCompile(`(\d+) files? changed, (\d+) insertions?\(\+\), (\d+) deletions?\(-\)`)

func (p *Parser) Parse(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	// Parsing for 'git status'
	if p.command == "status" {
		if strings.HasPrefix(line, "modified:") {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2611 "+line)
			p.summary.Metrics.Warnings++ // Treat modified files as something to watch
		} else if strings.HasPrefix(line, "new file:") {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u271A "+line)
		} else if strings.Contains(line, "branch") {
			p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\uf418 "+line)
		}
	}

	// Parsing for 'git diff' or 'git diff --stat'
	match := diffStatRegex.FindStringSubmatch(line)
	if len(match) > 0 {
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u231B "+line)
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.SummaryText = "Git " + p.command + " execution finished."
	if len(p.summary.ImportantEvents) > 0 {
		p.summary.SummaryText = "Git " + p.command + " summary: " + strings.Join(p.summary.ImportantEvents, ", ")
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
