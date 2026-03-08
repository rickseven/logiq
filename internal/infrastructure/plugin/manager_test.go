package plugin

import (
	"github.com/rickseven/logiq/internal/domain"
	"regexp"
	"testing"
)

// Mock plugin explicitly defining strict return arrays organically simulating actual plugins.
type MockPlugin struct{}

func (m *MockPlugin) Name() string { return "mock_plugin" }
func (m *MockPlugin) Parsers() []domain.Parser {
	return []domain.Parser{
		// Since flutterbuild/New isn't importable here natively without cycles during test phases,
		// We can just return nil array for unit-tests guaranteeing array mapping processes directly seamlessly.
	}
}
func (m *MockPlugin) DebugRules() []domain.SuggestionRule {
	return []domain.SuggestionRule{
		{
			Pattern:    regexp.MustCompile(`(?i)mocking rules`),
			ErrorType:  "mock_error",
			Suggestion: "Test fallback suggestions directly natively passed from plugin architecture bounds.",
		},
	}
}

func TestPluginRegistration(t *testing.T) {
	pluginObj := &MockPlugin{}
	Register(pluginObj)

	active := GetActivePlugins()

	if len(active) == 0 {
		t.Fatalf("Plugin failed to register successfully.")
	}

	if active[0].Name() != "mock_plugin" {
		t.Errorf("Expected 'mock_plugin', got '%s'", active[0].Name())
	}
}
