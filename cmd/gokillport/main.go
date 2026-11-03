package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/nopedal/gokillport/internal/finder"
	"github.com/nopedal/gokillport/internal/formatter"
	"github.com/nopedal/gokillport/internal/killer"
	"github.com/spf13/cobra"
)

var (
	// version is injected with -ldflags "-X main.version=...". It must not
	// have a function-call initializer, or the linker ignores -X.
	version string

	flagForce   bool
	flagDryRun  bool
	flagList    bool
	flagTimeout int
	flagQuiet   bool
	flagJSON    bool
)

func main() {
	root := &cobra.Command{
		Use:     "gokillport [ports...]",
		Short:   "Find and kill processes listening on given ports",
		Version: getVersion(),
		Args: func(cmd *cobra.Command, args []string) error {
			if flagList {
				return nil
			}
			if len(args) == 0 {
				return fmt.Errorf("at least one port number is required")
			}
			for _, arg := range args {
				if _, err := strconv.Atoi(arg); err != nil {
					return fmt.Errorf("invalid port number: %s", arg)
				}
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          run,
	}

	root.Flags().BoolVarP(&flagForce, "force", "f", false, "send SIGKILL instead of SIGTERM")
	root.Flags().BoolVarP(&flagDryRun, "dry-run", "n", false, "show what would be killed without killing")
	root.Flags().BoolVarP(&flagList, "list", "l", false, "list all processes bound to ports")
	root.Flags().IntVarP(&flagTimeout, "timeout", "t", 5, "seconds to wait before sending SIGKILL")
	root.Flags().BoolVarP(&flagQuiet, "quiet", "q", false, "suppress all output except errors")
	root.Flags().BoolVar(&flagJSON, "json", false, "output results as JSON")

	if err := root.Execute(); err != nil {
		if !flagQuiet {
			formatter.PrintError(os.Stderr, err.Error())
		}
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	f := finder.New()
	k := killer.New()
	w := os.Stdout

	if flagList {
		return runList(f, w)
	}

	exitCode := 0
	var results []killer.Result

	for _, arg := range args {
		port, _ := strconv.Atoi(arg)
		result := killPort(f, k, port, w)
		results = append(results, result)

		if result.Error != "" {
			switch {
			case result.Error == "not found":
				if exitCode == 0 {
					exitCode = 1
				}
			case result.Error == "permission denied":
				exitCode = 2
			case result.Error == "timeout":
				exitCode = 3
			default:
				if exitCode == 0 {
					exitCode = 1
				}
			}
		}
	}

	if flagJSON {
		formatter.PrintJSON(w, results)
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}

func killPort(f finder.PortFinder, k killer.ProcessKiller, port int, w *os.File) killer.Result {
	info, err := f.FindByPort(port)
	if err != nil {
		if !flagQuiet && !flagJSON {
			formatter.PrintError(w, fmt.Sprintf("port %d: %v", port, err))
		}
		return killer.Result{Port: port, Error: err.Error()}
	}

	if info == nil {
		if !flagQuiet && !flagJSON {
			formatter.PrintNotFound(w, port)
		}
		return killer.Result{Port: port, Error: "not found"}
	}

	signal := "SIGTERM"
	if flagForce {
		signal = "SIGKILL"
	}

	result := killer.Result{
		Port:    port,
		PID:     info.PID,
		Process: info.Process,
		Signal:  signal,
	}

	if flagDryRun {
		if !flagQuiet && !flagJSON {
			formatter.PrintKillLine(w, info, true)
		}
		result.Killed = false
		return result
	}

	// Send the signal
	err = k.Kill(info.PID, flagForce)
	if err != nil {
		if killer.IsProcessGone(err) {
			if !flagQuiet && !flagJSON {
				formatter.PrintProcessGone(w, info.PID, port)
			}
			result.Killed = true
			return result
		}
		if killer.IsPermissionError(err) {
			if !flagQuiet && !flagJSON {
				formatter.PrintPermissionError(w, port, info.PID)
			}
			result.Error = "permission denied"
			return result
		}
		if !flagQuiet && !flagJSON {
			formatter.PrintError(w, fmt.Sprintf("port %d: %v", port, err))
		}
		result.Error = err.Error()
		return result
	}

	// If not force, wait for graceful exit then escalate
	if !flagForce {
		timeout := time.Duration(flagTimeout) * time.Second
		if !k.WaitForExit(info.PID, timeout) {
			// Escalate to SIGKILL
			if !flagQuiet && !flagJSON {
				formatter.PrintTimeoutError(w, info.PID, port)
			}
			_ = k.Kill(info.PID, true)
			result.Signal = "SIGKILL (escalated)"

			// Brief wait after SIGKILL
			if !k.WaitForExit(info.PID, 2*time.Second) {
				result.Error = "timeout"
				return result
			}
		}
	}

	result.Killed = true
	if !flagQuiet && !flagJSON {
		formatter.PrintKillLine(w, info, false)
	}
	return result
}

func runList(f finder.PortFinder, w *os.File) error {
	procs, err := f.ListAll()
	if err != nil {
		return fmt.Errorf("listing ports: %w", err)
	}

	if flagJSON {
		return formatter.PrintJSON(w, procs)
	}

	formatter.PrintTable(w, procs)
	return nil
}

func getVersion() string {
	// Prefer ldflags-injected version at build time.
	if version != "" {
		return version
	}
	// Fall back to Go module version (set by `go install ...@v0.1.0`).
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
