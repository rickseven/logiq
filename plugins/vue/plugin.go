package vue

import (
	"github.com/rickseven/logiq/internal/app/parser/vite"
	"github.com/rickseven/logiq/internal/app/parser/vitest"
	"github.com/rickseven/logiq/internal/domain"
	"regexp"
)

type VuePlugin struct{}

// New instantiate the vue extension directly
func New() domain.Plugin {
	return &VuePlugin{}
}

func (p *VuePlugin) Name() string {
	return "vue"
}

func (p *VuePlugin) Parsers() []domain.Parser {
	return []domain.Parser{
		vite.NewParser(),
		vitest.NewParser(),
	}
}

func (p *VuePlugin) DebugRules() []domain.SuggestionRule {
	return []domain.SuggestionRule{
		{
			Pattern:    regexp.MustCompile(`(?i)vue compiler macro`),
			ErrorType:  "lint_error",
			Suggestion: "Ensure you are using defineProps and defineEmits correctly inside <script setup> block natively.",
		},
	}
}
