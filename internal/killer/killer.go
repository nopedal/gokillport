package killer

import (
	"errors"
	"fmt"
	"time"
)

// Common errors with semantic meaning for exit codes.
var (
	ErrNotFound   = errors.New("process not found")
	ErrPermission = errors.New("permission denied")
	ErrTimeout    = errors.New("process did not exit within timeout")
)

// Result represents the outcome of a kill operation.
type Result struct {
	Port    int    `json:"port"`
	PID     int    `json:"pid"`
	Process string `json:"process"`
	Killed  bool   `json:"killed"`
	Signal  string `json:"signal"`
	Error   string `json:"error,omitempty"`
}

// ProcessKiller kills processes by PID.
type ProcessKiller interface {
	// Kill sends a signal to the given PID. If force is true, sends SIGKILL,
	// otherwise SIGTERM.
	Kill(pid int, force bool) error

	// WaitForExit waits for the process to exit, returning true if it exited
	// within the timeout, false otherwise.
	WaitForExit(pid int, timeout time.Duration) bool

	// IsRunning checks if a process with the given PID is still running.
	IsRunning(pid int) bool
}

// New returns a ProcessKiller for the current platform.
func New() ProcessKiller {
	return newPlatformKiller()
}

// IsPermissionError checks if an error is a permission error.
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return errors.Is(err, ErrPermission) ||
		containsAny(msg, "operation not permitted", "access is denied", "permission denied")
}

// IsProcessGone checks if an error indicates the process no longer exists.
func IsProcessGone(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return containsAny(msg, "no such process", "process not found", "not found")
}

func containsAny(s string, substrs ...string) bool {
	lower := fmt.Sprintf("%s", s)
	for _, sub := range substrs {
		if len(lower) >= len(sub) {
			for i := 0; i <= len(lower)-len(sub); i++ {
				match := true
				for j := 0; j < len(sub); j++ {
					c := lower[i+j]
					sc := sub[j]
					// case-insensitive
					if c != sc && c != sc-32 && c != sc+32 {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}
