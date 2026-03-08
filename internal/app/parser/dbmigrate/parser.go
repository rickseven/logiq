package dbmigrate

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
	}
}

func (p *Parser) Tool() string { return p.tool }

func (p *Parser) Detect(cmd string, args []string) bool {
	tools := []string{"prisma", "migrate", "sequelize", "drizzle", "alembic", "knex"}
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
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2717 "+strings.TrimSpace(line))
	} else if strings.Contains(lower, "migrated") || strings.Contains(lower, "migration successfully") {
		p.summary.Status = "success"
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2611 "+strings.TrimSpace(line))
	} else if strings.Contains(line, "Applying migration") || strings.Contains(line, "Running migration") {
		p.summary.ImportantEvents = append(p.summary.ImportantEvents, "\u2192 "+strings.TrimSpace(line))
	}
}

func (p *Parser) Summary() domain.Summary {
	p.summary.Metrics.DurationSeconds = time.Since(p.start).Seconds()
	if p.summary.Status == "failure" {
		p.summary.SummaryText = "\u2717 Migration failed"
	} else {
		p.summary.SummaryText = "\u2713 DB Migration finished"
	}
	return p.summary
}

func (p *Parser) DetectFromContent(line string) bool {
	return false
}
