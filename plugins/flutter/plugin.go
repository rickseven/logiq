package flutter

import (
	"github.com/rickseven/logiq/internal/app/parser/flutterbuild"
	"github.com/rickseven/logiq/internal/app/parser/fluttertest"
	"github.com/rickseven/logiq/internal/domain"
	"regexp"
)

type FlutterPlugin struct{}

// New instantiate the flutter extension directly mapping parsers implicitly
func New() domain.Plugin {
	return &FlutterPlugin{}
}

func (p *FlutterPlugin) Name() string {
	return "flutter"
}

func (p *FlutterPlugin) Parsers() []domain.Parser {
	return []domain.Parser{
		flutterbuild.NewParser(),
		fluttertest.NewParser(),
	}
}

func (p *FlutterPlugin) DebugRules() []domain.SuggestionRule {
	return []domain.SuggestionRule{
		{
			Pattern:    regexp.MustCompile(`(?i)MissingPluginException`),
			ErrorType:  "exception",
			Suggestion: "Run 'flutter clean' and 'flutter pub get' to rebuild generated plugin wrappers natively.",
		},
	}
}
