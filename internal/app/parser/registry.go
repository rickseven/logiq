package parser

import (
	"github.com/rickseven/logiq/internal/app/parser/buildtool"
	"github.com/rickseven/logiq/internal/app/parser/cargo"
	"github.com/rickseven/logiq/internal/app/parser/cypress"
	"github.com/rickseven/logiq/internal/app/parser/dart"
	"github.com/rickseven/logiq/internal/app/parser/dbmigrate"
	"github.com/rickseven/logiq/internal/app/parser/dotnet"
	"github.com/rickseven/logiq/internal/app/parser/eslint"
	"github.com/rickseven/logiq/internal/app/parser/fallback"
	"github.com/rickseven/logiq/internal/app/parser/flutterbuild"
	"github.com/rickseven/logiq/internal/app/parser/flutterrun"
	"github.com/rickseven/logiq/internal/app/parser/fluttertest"
	"github.com/rickseven/logiq/internal/app/parser/git"
	"github.com/rickseven/logiq/internal/app/parser/golangcilint"
	"github.com/rickseven/logiq/internal/app/parser/gotest"
	"github.com/rickseven/logiq/internal/app/parser/gradle"
	"github.com/rickseven/logiq/internal/app/parser/jest"
	"github.com/rickseven/logiq/internal/app/parser/linter"
	"github.com/rickseven/logiq/internal/app/parser/npm"
	"github.com/rickseven/logiq/internal/app/parser/playwright"
	"github.com/rickseven/logiq/internal/app/parser/pnpm"
	"github.com/rickseven/logiq/internal/app/parser/pytest"
	"github.com/rickseven/logiq/internal/app/parser/typescript"
	"github.com/rickseven/logiq/internal/app/parser/vite"
	"github.com/rickseven/logiq/internal/app/parser/vitest"
	"github.com/rickseven/logiq/internal/app/parser/yarn"
	"github.com/rickseven/logiq/internal/domain"
)

// ParserFactory is a function that instantiates a new parser
type ParserFactory func() domain.Parser

var pluginParsers []ParserFactory

// RegisterPluginParser securely adds external compiled parsers mapping to priorities organically
func RegisterPluginParser(pf ParserFactory) {
	pluginParsers = append(pluginParsers, pf)
}

// GetParsers returns all registered parsers explicitly mapped
func GetParsers() []ParserFactory {
	base := []ParserFactory{
		// Test frameworks have high priority
		func() domain.Parser { return vitest.NewParser() },
		func() domain.Parser { return jest.NewParser() },
		func() domain.Parser { return playwright.NewParser() },
		func() domain.Parser { return cypress.NewParser() },
		func() domain.Parser { return pytest.NewParser() },
		func() domain.Parser { return gotest.NewParser() },
		func() domain.Parser { return fluttertest.NewParser() },

		// Type Checking
		func() domain.Parser { return typescript.NewParser() },
		func() domain.Parser { return dart.NewParser() },

		// Build/Lint frameworks
		func() domain.Parser { return vite.NewParser() },
		func() domain.Parser { return buildtool.NewParser() },
		func() domain.Parser { return eslint.NewParser() },
		func() domain.Parser { return linter.NewParser() },
		func() domain.Parser { return golangcilint.NewParser() },
		func() domain.Parser { return flutterbuild.NewParser() },
		func() domain.Parser { return flutterrun.NewParser() },
		func() domain.Parser { return gradle.NewParser() },
		func() domain.Parser { return dotnet.NewParser() },
		func() domain.Parser { return dbmigrate.NewParser() },

		// General purpose dependency managers (lower priority because they often wrap others)
		func() domain.Parser { return npm.NewParser() },
		func() domain.Parser { return pnpm.NewParser() },
		func() domain.Parser { return yarn.NewParser() },
		func() domain.Parser { return cargo.NewParser() },
		func() domain.Parser { return git.NewParser() },

		// Fallback handles anything else
		func() domain.Parser { return fallback.NewParser() },
	}

	all := make([]ParserFactory, 0, len(pluginParsers)+len(base))
	all = append(all, pluginParsers...)
	all = append(all, base...)
	return all
}
