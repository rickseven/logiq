package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rickseven/logiq/internal/app"
	"github.com/rickseven/logiq/internal/domain"
	"github.com/rickseven/logiq/internal/infrastructure/artifacts"
	"github.com/rickseven/logiq/internal/infrastructure/config"
	"github.com/rickseven/logiq/internal/infrastructure/observability/logger"
	"github.com/rickseven/logiq/internal/infrastructure/plugin"
	"github.com/rickseven/logiq/internal/infrastructure/runtime"
	"github.com/rickseven/logiq/internal/interfaces/cli/output"
	"github.com/rickseven/logiq/internal/interfaces/mcp"
	"github.com/rickseven/logiq/plugins/flutter"
	"github.com/rickseven/logiq/plugins/vue"
)

// Local Version variable removed, using domain.GetVersion()

func main() {
	cfg := config.LoadConfig()

	mode := flag.String("mode", "human", "Output mode: human, json, agent")
	saveArtifacts := flag.Bool("save-artifacts", false, "Save output as structured artifact")
	timeout := flag.Int("timeout", cfg.Timeout, "Timeout in seconds")
	debug := flag.Bool("debug", cfg.Debug, "Enable debug mode")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: logiq run <command> [args...]\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	logger.InitLogger(*debug)

	cmdRunner := runtime.NewRunner()
	appService := app.NewService(cmdRunner)

	subcmd := args[0]
	if subcmd == "version" || subcmd == "-v" || subcmd == "--version" {
		fmt.Printf("LogIQ version %s\n", domain.GetVersion())
		return
	} else if subcmd == "analyze" {
		if len(args) < 2 {
			fmt.Println("Usage: logiq analyze <logs>")
			os.Exit(1)
		}
		res, _ := appService.ExecuteCommand(context.Background(), "analyze", []string{"--raw", strings.Join(args[1:], " ")}, *debug)
		output.Print(*mode, res)
		return
	} else if subcmd == "explain" {
		if len(args) < 2 {
			fmt.Println("Usage: logiq explain <command>")
			os.Exit(1)
		}
		res := appService.Explain(args[1:])
		output.PrintExplain(*mode, res)
		return
	} else if subcmd == "doctor" {
		res := appService.Diagnose()
		output.PrintDoctor(*mode, res)
		return
	} else if subcmd == "trace" {
		res := appService.History()
		output.PrintTrace(*mode, res)
		return
	} else if subcmd == "plugins" {
		plugin.HandleCLI(args[1:])
		return
	} else if subcmd == "mcp" {
		server := mcp.NewStdioServer()
		server.Start()
		return
	} else if subcmd != "run" || len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	// Load predefined simulated plugins statically mapped (dynamic .so limits fallback)
	plugin.Register(flutter.New())
	plugin.Register(vue.New())

	commandArgs := args[1:]

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	result, err := appService.ExecuteCommand(ctx, commandArgs[0], commandArgs[1:], *debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start command: %v\n", err)
		os.Exit(1)
	}

	// Artifact Generation
	if *saveArtifacts {
		var fileName string
		switch result.Tool {
		case "fallback", "":
			fileName = "execution_summary.json"
		case "gotest", "pytest", "jest":
			fileName = "test_results.json"
		case "cargo":
			fileName = "build_summary.json"
		default:
			fileName = fmt.Sprintf("%s_summary.json", result.Tool)
		}

		if err := artifacts.WriteArtifact(fileName, result); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write artifact: %v\n", err)
		} else if *debug {
			fmt.Fprintf(os.Stderr, "[DEBUG LOGIQ] Exported artifacts to .logiq/artifacts\n")
		}
	}

	// Record execution locally to trace log
	cmdLine := strings.Join(commandArgs, " ")
	appService.RecordTrace(result.ExecutionID, cmdLine, result.Status, result.Summary)

	// Print Output
	output.Print(*mode, result)

	if result.Status == "failure" {
		os.Exit(1) // Return original command failure exit code ideally, but 1 is fine here.
	}
}
