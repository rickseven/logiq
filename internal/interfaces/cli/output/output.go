package output

import (
	"encoding/json"
	"fmt"

	"github.com/rickseven/logiq/internal/domain"
)

// Print formats and prints the output based on mode (human, agent, json)
func Print(mode string, out domain.StructuredOutput) {
	switch mode {
	case "json":
		bytes, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(bytes))
	case "agent":
		agentOutput := map[string]interface{}{
			"compressed_context": out.CompressedContext,
			"metrics":            out.Metrics,
		}

		if out.Status == "failure" && out.RawLogPath != "" {
			agentOutput["raw_log_fallback"] = out.RawLogPath
		}

		if out.ErrorIntel != nil {
			agentOutput["errors"] = out.ErrorIntel
		} else {
			agentOutput["errors"] = nil
		}

		artifacts := map[string]string{}
		if out.Metrics.ArtifactSize != "" {
			artifacts["artifact_size"] = out.Metrics.ArtifactSize
		}
		if out.Metrics.BundleSize != "" {
			artifacts["bundle_size"] = out.Metrics.BundleSize
		}
		agentOutput["artifacts"] = artifacts

		bytes, _ := json.Marshal(agentOutput)
		fmt.Println(string(bytes))
	case "human":
		fallthrough
	default:
		icon := "✅"
		if out.Status != "success" {
			icon = "❌"
		}

		fmt.Printf("%s %s\n", icon, out.Summary)

		if len(out.Suggestions) > 0 {
			fmt.Println("\nSuggested Fix:")
			for _, s := range out.Suggestions {
				fmt.Printf("💡 %s\n", s)
			}
		}

		if out.Metrics.DurationSeconds > 0 {
			fmt.Printf("\n⏱️  Duration: %.2fs\n", out.Metrics.DurationSeconds)
		}
		if out.Metrics.TestsPassed > 0 || out.Metrics.TestsFailed > 0 {
			fmt.Printf("   Passed: %d, Failed: %d\n", out.Metrics.TestsPassed, out.Metrics.TestsFailed)
		}
		if len(out.ImportantEvents) > 0 {
			fmt.Println("\nImportant Events:")
			for _, e := range out.ImportantEvents {
				fmt.Printf(" > %s\n", e)
			}
		}

		if out.Metrics.OriginalBytes > 0 {
			fmt.Printf("\n📉 Token/Context Savings: %.1f%% (%d B ➔ %d B)\n", out.Metrics.SavingsPercentage, out.Metrics.OriginalBytes, out.Metrics.CompressedBytes)
		}

		if out.Status == "failure" && out.RawLogPath != "" {
			fmt.Printf("\n[Tee Mode Fallback: Raw log saved to %s]\n", out.RawLogPath)
		}
	}
}
