package runtime

import (
	"bufio"
	"context"
	"io"
	"os/exec"
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

func NewRunner() Runner {
	return &DefaultRunner{}
}

func (r *DefaultRunner) Run(ctx context.Context, cmdName string, args []string) (<-chan string, <-chan int, *domain.ResourceMetrics, error) {
	var cmd *exec.Cmd
	metrics := &domain.ResourceMetrics{}

	if stdruntime.GOOS == "windows" {
		var sb strings.Builder
		sb.WriteString(quoteArg(cmdName))
		for _, arg := range args {
			sb.WriteString(" ")
			sb.WriteString(quoteArg(arg))
		}
		fullCommand := sb.String()
		cmd = exec.CommandContext(ctx, "cmd", "/C", fullCommand)
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
