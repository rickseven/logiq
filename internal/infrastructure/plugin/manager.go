package plugin

import (
	"fmt"
	"os"

	"github.com/rickseven/logiq/internal/app/debugassist"
	"github.com/rickseven/logiq/internal/app/parser"
	"github.com/rickseven/logiq/internal/domain"
)

var activePlugins []Plugin

// Register loads a plugin into the engine and injects its components
func Register(p Plugin) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Warning: plugin %s panicked during registration: %v\n", p.Name(), r)
		}
	}()

	activePlugins = append(activePlugins, p)

	// Register parsers securely. We wrap the exact instance into a factory.
	for _, pInst := range p.Parsers() {
		inst := pInst
		parser.RegisterPluginParser(func() domain.Parser {
			return inst
		})
	}

	// Register debug rules
	for _, rule := range p.DebugRules() {
		debugassist.RegisterRule(rule)
	}
}

// GetActivePlugins returns a list of loaded plugins
func GetActivePlugins() []Plugin {
	return activePlugins
}
