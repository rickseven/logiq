package plugin

import (
	"github.com/rickseven/logiq/internal/domain"
)

// Plugin interface allows external modules to extend LogIQ without modifying core code.
// DEPRECATED: Please use domain.Plugin instead.
// For backwards compatibility during refactoring
type Plugin = domain.Plugin
