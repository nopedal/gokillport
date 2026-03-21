package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nopedal/gokillport/internal/finder"
	"github.com/nopedal/gokillport/internal/killer"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))
)

// PrintKillLine prints what is about to be (or was) killed.
func PrintKillLine(w io.Writer, info *finder.ProcessInfo, dryRun bool) {
	if dryRun {
		fmt.Fprintf(w, "%s %s %s %s\n",
			warnStyle.Render("would kill"),
			headerStyle.Render(info.Process),
			dimStyle.Render(fmt.Sprintf("(pid %d)", info.PID)),
			dimStyle.Render(fmt.Sprintf("on port %d", info.Port)),
		)
	} else {
		fmt.Fprintf(w, "%s %s %s %s\n",
			successStyle.Render("killed"),
			headerStyle.Render(info.Process),
			dimStyle.Render(fmt.Sprintf("(pid %d)", info.PID)),
			dimStyle.Render(fmt.Sprintf("on port %d", info.Port)),
		)
	}
}

// PrintNotFound prints a message when no process is found on a port.
func PrintNotFound(w io.Writer, port int) {
	fmt.Fprintf(w, "%s no process found on port %d\n",
		errorStyle.Render("error:"),
		port,
	)
}

// PrintPermissionError prints a permission error with sudo hint.
func PrintPermissionError(w io.Writer, port int, pid int) {
	fmt.Fprintf(w, "%s permission denied killing pid %d on port %d\n",
		errorStyle.Render("error:"),
		pid, port,
	)
	fmt.Fprintf(w, "%s\n",
		dimStyle.Render("  hint: rerun with sudo"),
	)
}

// PrintTimeoutError prints a timeout message.
func PrintTimeoutError(w io.Writer, pid int, port int) {
	fmt.Fprintf(w, "%s pid %d on port %d did not exit in time, sent SIGKILL\n",
		warnStyle.Render("timeout:"),
		pid, port,
	)
}

// PrintProcessGone prints a message when the process disappeared.
func PrintProcessGone(w io.Writer, pid int, port int) {
	fmt.Fprintf(w, "%s pid %d on port %d already exited\n",
		dimStyle.Render("info:"),
		pid, port,
	)
}

// PrintError prints a generic error.
func PrintError(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s %s\n", errorStyle.Render("error:"), msg)
}

// PrintTable prints a list of ProcessInfo as a formatted table.
func PrintTable(w io.Writer, procs []ProcessInfo) {
	if len(procs) == 0 {
		fmt.Fprintln(w, dimStyle.Render("no listening processes found"))
		return
	}

	// Column widths
	portW, pidW, procW, userW := 6, 8, 20, 15
	for _, p := range procs {
		if l := len(fmt.Sprintf("%d", p.Port)); l > portW {
			portW = l
		}
		if l := len(fmt.Sprintf("%d", p.PID)); l > pidW {
			pidW = l
		}
		if l := len(p.Process); l > procW {
			procW = l
		}
		if l := len(p.User); l > userW {
			userW = l
		}
	}

	// Header
	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s",
		portW, "PORT", pidW, "PID", procW, "PROCESS", userW, "USER")
	fmt.Fprintln(w, headerStyle.Render(header))
	fmt.Fprintln(w, dimStyle.Render(strings.Repeat("─", portW+pidW+procW+userW+6)))

	// Rows
	for _, p := range procs {
		fmt.Fprintf(w, "%-*d  %-*d  %-*s  %-*s\n",
			portW, p.Port, pidW, p.PID, procW, p.Process, userW, p.User)
	}
}

// PrintJSON outputs results as JSON.
func PrintJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ProcessInfo is a re-export for table formatting convenience.
type ProcessInfo = finder.ProcessInfo

// KillResult is a re-export for JSON output.
type KillResult = killer.Result
