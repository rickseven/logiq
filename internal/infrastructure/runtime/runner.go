package runtime

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"regexp"
	stdruntime "runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/rickseven/logiq/internal/domain"
	"github.com/shirou/gopsutil/v3/process"
)

func sanitizeUTF8(input string) string {
	if utf8.ValidString(input) {
		return input
	}
	return strings.ToValidUTF8(input, "")
}

type Runner interface {
	Run(ctx context.Context, cmd string, args []string) (<-chan string, <-chan int, *domain.ResourceMetrics, error)
}

type DefaultRunner struct{}

var powershellCmdletPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*-[A-Za-z][A-Za-z0-9]*$`)

// powershellCmdletInTextPattern matches PowerShell Verb-Noun cmdlet patterns
// anywhere in a command string (e.g. Select-String, Get-ChildItem, Invoke-WebRequest).
// This covers ALL PowerShell cmdlets without needing a hardcoded list.
// Standard PS approved verbs: Get, Set, New, Remove, Add, Clear, Close, Copy, Enter,
// Exit, Find, Format, Hide, Import, Export, Install, Invoke, Join, Lock, Measure,
// Move, Open, Out, Pop, Push, Redo, Register, Rename, Repair, Request, Reset,
// Resize, Resolve, Restart, Restore, Resume, Save, Search, Select, Send, Show,
// Skip, Split, Start, Step, Stop, Submit, Suspend, Switch, Sync, Test, Trace,
// Undo, Uninstall, Unlock, Unprotect, Unregister, Update, Wait, Watch, Write, etc.
var powershellCmdletInTextPattern = regexp.MustCompile(
	`(?i)\b(Get|Set|New|Remove|Add|Clear|Close|Copy|Enter|Exit|Find|Format|Hide|` +
		`Import|Export|Install|Invoke|Join|Lock|Measure|Move|Open|Out|Pop|Push|Redo|` +
		`Register|Rename|Repair|Request|Reset|Resize|Resolve|Restart|Restore|Resume|` +
		`Save|Search|Select|Send|Show|Skip|Split|Start|Step|Stop|Submit|Suspend|` +
		`Switch|Sync|Test|Trace|Undo|Uninstall|Unlock|Unprotect|Unregister|Update|` +
		`Wait|Watch|Write|Compare|Complete|Compress|Confirm|Connect|Convert|ConvertFrom|` +
		`ConvertTo|Debug|Deny|Disable|Disconnect|Dismount|Edit|Enable|Expand|Grant|` +
		`Group|Initialize|Limit|Merge|Mount|Optimize|Ping|Protect|Publish|Receive|` +
		`Revoke|Sort|Unpublish|Use|Where|ForEach)-[A-Z][A-Za-z0-9]+\b`,
)

// Patterns that indicate PowerShell-specific syntax (variables, operators, sub-expressions, type casts)
var powershellSyntaxPattern = regexp.MustCompile(
	`(?i)` +
		// PS variables and environment
		`\$env:|\$\{[^}]+\}|\$\([^)]+\)|` +
		// PS array/hashtable literals
		`@\(|@\{|` +
		// PS comparison/logical operators
		`\s-(?:eq|ne|gt|lt|ge|le|like|notlike|match|notmatch|contains|notcontains|in|notin|is|isnot|as|replace|split|join|band|bor|bxor|bnot|shl|shr|and|or|not|f)\s|` +
		// PS type cast syntax [Type]::
		`\[[A-Za-z]+(?:\.[A-Za-z]+)*\]\s*::` +
		// PS pipeline variable $_
		`|\$_\.`,
)

func NewRunner() Runner {
	return &DefaultRunner{}
}

func (r *DefaultRunner) Run(ctx context.Context, cmdName string, args []string) (<-chan string, <-chan int, *domain.ResourceMetrics, error) {
	var cmd *exec.Cmd
	metrics := &domain.ResourceMetrics{}

	if stdruntime.GOOS == "windows" {
		if strings.EqualFold(cmdName, "cmd") || strings.EqualFold(cmdName, "cmd.exe") {
			// Avoid wrapping cmd inside another cmd /C because it can alter /c payload parsing.
			cmd = exec.CommandContext(ctx, "cmd", args...)
		} else if strings.EqualFold(cmdName, "powershell") || strings.EqualFold(cmdName, "powershell.exe") ||
			strings.EqualFold(cmdName, "pwsh") || strings.EqualFold(cmdName, "pwsh.exe") {
			// Explicitly invoked PowerShell — pass args through directly
			cmd = exec.CommandContext(ctx, cmdName, args...)
		} else if looksLikePowerShellCmdlet(cmdName) && hasPowerShellInstalled() {
			psExe := getPowerShellExe()
			fullCommand := buildPowerShellCommand(cmdName, args)
			cmd = exec.CommandContext(ctx, psExe, "-NoProfile", "-NonInteractive", "-Command", fullCommand)
		} else if containsPowerShellSyntax(cmdName, args) && hasPowerShellInstalled() {
			// The command contains PowerShell-specific syntax embedded in args
			// (e.g. "cd C:\path; Select-String ..." or "$env:VAR")
			// Route through PowerShell instead of cmd.exe
			psExe := getPowerShellExe()
			var fullCommand string
			if len(args) == 0 && containsShellOperators(cmdName) {
				// cmdName is the entire pipeline/chained expression (e.g. "Get-ChildItem . | Measure-Object").
				// Pass it as-is to avoid quoting the shell operators inside it.
				fullCommand = cmdName
			} else {
				fullCommand = buildPowerShellCommand(cmdName, args)
			}
			cmd = exec.CommandContext(ctx, psExe, "-NoProfile", "-NonInteractive", "-Command", fullCommand)
		} else {
			fullCommand := buildCmdCommand(cmdName, args)
			cmd = exec.CommandContext(ctx, "cmd", "/C", fullCommand)
		}
	} else {
		cmd = exec.CommandContext(ctx, cmdName, args...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, nil, err
	}

	outChan := make(chan string, 1000)
	exitChan := make(chan int, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	streamToChan := func(reader io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(reader)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			outChan <- sanitizeUTF8(scanner.Text())
		}
	}

	go streamToChan(stdout)
	go streamToChan(stderr)

	// Resource Monitoring Loop
	monitorCtx, monitorCancel := context.WithCancel(context.Background())
	go func() {
		defer monitorCancel()
		proc, err := process.NewProcess(int32(cmd.Process.Pid))
		if err != nil {
			return
		}

		var cpuSamples []float64
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-monitorCtx.Done():
				if len(cpuSamples) > 0 {
					var total float64
					for _, s := range cpuSamples {
						total += s
					}
					metrics.AvgCPUPercent = total / float64(len(cpuSamples))
				}
				return
			case <-ticker.C:
				if mem, err := proc.MemoryInfo(); err == nil {
					ramMB := float64(mem.RSS) / 1024 / 1024
					if ramMB > metrics.MaxRAMMB {
						metrics.MaxRAMMB = ramMB
					}
				}
				if cpu, err := proc.CPUPercent(); err == nil && cpu > 0 {
					cpuSamples = append(cpuSamples, cpu)
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		waitErr := cmd.Wait()
		monitorCancel() // Stop monitoring
		close(outChan)

		exitCode := 0
		if waitErr != nil {
			if exitError, ok := waitErr.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = 1
			}
		}
		exitChan <- exitCode
		close(exitChan)
	}()

	return outChan, exitChan, metrics, nil
}

func hasPowerShellInstalled() bool {
	// Check for PowerShell Core (pwsh) first, then Windows PowerShell (powershell)
	_, err := exec.LookPath("pwsh")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("powershell")
	return err == nil
}

// getPowerShellExe returns the best available PowerShell executable.
// Prefers pwsh (PowerShell Core / 7+) over powershell (Windows PowerShell 5.1).
func getPowerShellExe() string {
	if _, err := exec.LookPath("pwsh"); err == nil {
		return "pwsh"
	}
	return "powershell"
}

func looksLikePowerShellCmdlet(cmdName string) bool {
	return powershellCmdletPattern.MatchString(cmdName)
}

// containsShellOperators reports whether s contains shell pipeline or chaining operators.
// Used to detect when cmdName is an entire pipeline expression rather than a bare command name.
func containsShellOperators(s string) bool {
	return strings.ContainsAny(s, "|;&<>")
}

// containsPowerShellSyntax checks if the full command (name + args) contains
// PowerShell-specific syntax that cmd.exe cannot execute.
func containsPowerShellSyntax(cmdName string, args []string) bool {
	// Build full command string for pattern matching
	full := cmdName + " " + strings.Join(args, " ")

	// Check for PowerShell Verb-Noun cmdlet pattern anywhere in the command
	if powershellCmdletInTextPattern.MatchString(full) {
		return true
	}

	// Check for PowerShell variable/operator/syntax patterns
	if powershellSyntaxPattern.MatchString(full) {
		return true
	}

	return false
}

func buildCmdCommand(cmdName string, args []string) string {
	var sb strings.Builder
	sb.WriteString(quoteArg(cmdName))
	for _, arg := range args {
		sb.WriteString(" ")
		sb.WriteString(quoteArg(arg))
	}
	return sb.String()
}

func buildPowerShellCommand(cmdName string, args []string) string {
	var sb strings.Builder
	sb.WriteString(quotePowerShellArg(cmdName))
	for _, arg := range args {
		sb.WriteString(" ")
		sb.WriteString(quotePowerShellArg(arg))
	}
	return sb.String()
}

func quoteArg(arg string) string {
	// Shell operators should not be quoted
	operators := map[string]bool{"&&": true, "||": true, ";": true, "|": true, ">": true, ">>": true, "<": true}
	if operators[arg] {
		return arg
	}

	// If it doesn't have spaces or special chars, don't quote
	if !strings.ContainsAny(arg, " \t\n\r&|;<>^(){}") {
		return arg
	}

	// Clean double quotes for cmd.exe
	return "\"" + strings.ReplaceAll(arg, "\"", "\"\"") + "\""
}

func quotePowerShellArg(arg string) string {
	operators := map[string]bool{"&&": true, "||": true, ";": true, "|": true, ">": true, ">>": true, "<": true}
	if operators[arg] {
		return arg
	}

	if !strings.ContainsAny(arg, " \t\n\r&|;<>^(){}[]$`\"'") {
		return arg
	}

	return "'" + strings.ReplaceAll(arg, "'", "''") + "'"
}
