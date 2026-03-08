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
	} else if strings.Contains(lower, "flutter build") {
		result.Type = "build"
		result.Tool = "flutter"
		result.Description = "Builds the Flutter application for the target platform."
	} else if strings.Contains(lower, "flutter test") {
		result.Type = "test"
		result.Tool = "flutter"
		result.Description = "Runs unit or widget tests for the Flutter project."

		// 4. VCS
	} else if strings.HasPrefix(lower, "git ") {
		result.Type = "vcs"
		result.Tool = "git"
		result.Description = "A version control command to manage source code history."
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
