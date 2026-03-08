package detector

import (
	"regexp"
	"strings"

	"github.com/rickseven/logiq/internal/app/parser"
	"github.com/rickseven/logiq/internal/domain"
)

type Detector interface {
	Detect(cmd string, args []string, logs []string) domain.Parser
}

type DefaultDetector struct{}

func NewDetector() Detector {
	return &DefaultDetector{}
}

// decompositionRegex identifies common shell command separators
var decompositionRegex = regexp.MustCompile(`\s*(&&|;|\|\||\|)\s*`)

func (d *DefaultDetector) Detect(cmd string, args []string, logs []string) domain.Parser {
	parsers := parser.GetParsers()

	// 1. Decomposition for chaining (Support for any command chain)
	fullCmdLine := cmd + " " + strings.Join(args, " ")
	parts := decompositionRegex.Split(fullCmdLine, -1)

	// Try each part of the chain, from last to first
	// (Last command is usually the one that produces the output we want to parse)
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}

		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}

		subCmd := fields[0]
		subArgs := fields[1:]

		for _, pFactory := range parsers {
			p := pFactory()
			if p.Tool() == "fallback" {
				continue
			}
			if p.Detect(subCmd, subArgs) {
				return p
			}
		}
	}

	// 2. Try content-based detection if command decomposition fails (Smart Fallback)
	// This is naturally good for chaining because logs contain output from all parts
	for _, pFactory := range parsers {
		p := pFactory()
		if p.Tool() == "fallback" {
			continue
		}
		for _, line := range logs {
			if p.DetectFromContent(line) {
				return p
			}
		}
	}

	// 3. Final Fallback
	return parser.GetParsers()[len(parser.GetParsers())-1]()
}
