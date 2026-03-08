package domain

import (
	"runtime/debug"
)

// Version is the current version of LogIQ.
// It can be overridden at build time using ldflags:
// -ldflags="-X github.com/rickseven/logiq/internal/domain.Version=v1.0.0"
var Version = "1.0.0"

// GetVersion returns the version of the current build.
// If built via 'go install' or with module info, it tries to read the version automatically.
func GetVersion() string {
	if Version != "dev" {
		return Version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}

	return Version
}
