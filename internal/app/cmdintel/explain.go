package cmdintel

import (
	"fmt"
	"strings"

	"github.com/rickseven/logiq/internal/domain"
	"github.com/rickseven/logiq/internal/infrastructure/cache"
)

// GenerateExplainResult processes arguments to generate structured explanations
func GenerateExplainResult(args []string) domain.ExplainResult {
	cmdStr := strings.TrimSpace(strings.Join(args, " "))

	if cached, found := cache.GetExplain(cmdStr); found {
		return cached
	}

	var result domain.ExplainResult

	// Check for chaining
	if strings.Contains(cmdStr, " && ") || strings.Contains(cmdStr, " ; ") || strings.Contains(cmdStr, " || ") {
		result = explainChained(cmdStr)
	} else {
		result = explainSingle(cmdStr)
	}

	cache.SetExplain(cmdStr, result)
	return result
}

func explainSingle(cmdStr string) domain.ExplainResult {
	result := domain.ExplainResult{
		Command:     cmdStr,
		Type:        "unknown",
		Tool:        "unknown",
		Description: "Unknown command",
	}

	lower := strings.ToLower(cmdStr)
	fields := strings.Fields(cmdStr)

	// 1. High-priority specific tools
	if strings.Contains(lower, "vitest") {
		result.Type = "test"
		result.Tool = "vitest"
		if strings.Contains(lower, "--coverage") {
			result.Description = "Runs Vitest tests and generates a code coverage report to see how much of the code is tested."
		} else {
			result.Description = "Executes unit and widget tests using the Vitest test runner, highly optimized for Vite-based projects."
		}
	} else if strings.Contains(lower, "vue-tsc") || (strings.Contains(lower, "tsc") && !strings.Contains(lower, "vitest")) {
		result.Type = "typecheck"
		result.Tool = "typescript"
		result.Description = "Performs static type checking on the codebase to ensure type safety."
	} else if strings.Contains(lower, "npm run build") || strings.Contains(lower, "vite build") {
		result.Type = "build"
		result.Tool = "vite"
		result.Description = "Compiles the application for production. In modern Vue projects, this typically uses Vite/Rollup to bundle assets."
	} else if strings.Contains(lower, "npm run dev") || strings.Contains(lower, "npm start") || (strings.Contains(lower, "vite") && !strings.Contains(lower, "build")) {
		result.Type = "dev"
		result.Tool = "vite"
		result.Description = "Starts the development server with hot module replacement (HMR)."

		// 2. Generic package managers
	} else if strings.Contains(lower, "npx ") {
		result.Type = "exec"
		result.Tool = "npx"
		result.Description = "Executes a package-based command without having to install it globally."
	} else if strings.Contains(lower, "npm run test") || strings.Contains(lower, "npm test") {
		result.Type = "test"
		result.Tool = "npm"
		result.Description = "Executes the test suite defined in package.json."
	} else if strings.Contains(lower, "npm install") || strings.Contains(lower, "npm i") {
		result.Type = "install"
		result.Tool = "npm"
		result.Description = "Installs project dependencies defined in package.json."

		// 3. Other frameworks
	} else if strings.HasPrefix(lower, "flutter ") {
		return explainFlutterCommand(cmdStr, fields)

		// 4. VCS
	} else if strings.HasPrefix(lower, "git ") {
		result.Type = "vcs"
		result.Tool = "git"
		result.Description = "A version control command to manage source code history."
	}

	return result
}

func explainFlutterCommand(cmdStr string, fields []string) domain.ExplainResult {
	result := domain.ExplainResult{
		Command: cmdStr,
		Type:    "flutter",
		Tool:    "flutter",
	}

	if len(fields) < 2 {
		result.Description = "Runs a Flutter CLI command."
		return result
	}

	subcmd := strings.ToLower(fields[1])
	args := fields[2:]
	joined := " " + strings.ToLower(strings.Join(args, " ")) + " "

	findValue := func(flag string) string {
		for i := 0; i < len(args)-1; i++ {
			if strings.EqualFold(args[i], flag) {
				return args[i+1]
			}
		}
		return ""
	}

	mode := ""
	if strings.Contains(joined, " --debug ") {
		mode = "debug"
	} else if strings.Contains(joined, " --profile ") {
		mode = "profile"
	} else if strings.Contains(joined, " --release ") {
		mode = "release"
	}

	verbose := strings.Contains(joined, " -v ") || strings.Contains(joined, " --verbose ")
	flavor := findValue("--flavor")
	target := findValue("-t")
	if target == "" {
		target = findValue("--target")
	}

	details := make([]string, 0, 5)
	if mode != "" {
		details = append(details, "mode "+mode)
	}
	if flavor != "" {
		details = append(details, "flavor "+flavor)
	}
	if target != "" {
		details = append(details, "entrypoint "+target)
	}
	if strings.Contains(joined, " --dart-define") {
		details = append(details, "custom runtime defines")
	}
	if verbose {
		details = append(details, "verbose logs")
	}

	suffix := ""
	if len(details) > 0 {
		suffix = " with " + strings.Join(details, ", ")
	}

	switch subcmd {
	case "build":
		platform := "target platform"
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			platform = args[0]
		}
		result.Type = "build"
		result.Description = "Builds the Flutter app for " + platform + suffix + "."
	case "test":
		result.Type = "test"
		result.Description = "Runs Flutter tests" + suffix + "."
	case "run":
		result.Type = "run"
		result.Description = "Builds and launches the Flutter app on a connected device/emulator" + suffix + "."
	case "analyze":
		result.Type = "lint"
		result.Description = "Runs static analysis on Flutter/Dart source code to detect issues before build/runtime." + suffix
	case "pub":
		result.Type = "dependency"
		result.Description = "Runs Flutter pub package management command" + suffix + "."
	default:
		result.Type = "flutter"
		result.Description = "Runs Flutter command '" + subcmd + "'" + suffix + "."
	}

	return result
}

func explainChained(cmdStr string) domain.ExplainResult {
	// Simple split by known separators
	separators := []string{" && ", " ; ", " || "}
	var parts []string

	// Default to && for now as it's the most common
	currentParts := []string{cmdStr}
	for _, sep := range separators {
		var nextParts []string
		for _, p := range currentParts {
			nextParts = append(nextParts, strings.Split(p, sep)...)
		}
		currentParts = nextParts
	}
	parts = currentParts

	var descriptions []string
	mainType := "composite"
	mainTool := "multitool"

	for i, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		exp := explainSingle(p)
		descriptions = append(descriptions, fmt.Sprintf("%d. **%s**: %s", i+1, p, exp.Description))
	}

	return domain.ExplainResult{
		Command:     cmdStr,
		Type:        mainType,
		Tool:        mainTool,
		Description: "This is a chained command execution:\n\n" + strings.Join(descriptions, "\n\n"),
	}
}
