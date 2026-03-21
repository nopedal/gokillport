//go:build windows

package killer

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type windowsKiller struct{}

func newPlatformKiller() ProcessKiller {
	return &windowsKiller{}
}

func (k *windowsKiller) Kill(pid int, force bool) error {
	// On Windows, taskkill /T kills process tree, /F forces
	args := []string{"/PID", fmt.Sprintf("%d", pid)}
	if force {
		args = append(args, "/F")
	}
	args = append(args, "/T")

	cmd := exec.Command("taskkill", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(out)
		if IsPermissionError(fmt.Errorf("%s", msg)) {
			return fmt.Errorf("%w: %s", ErrPermission, msg)
		}
		if IsProcessGone(fmt.Errorf("%s", msg)) {
			return fmt.Errorf("%w: pid %d", ErrNotFound, pid)
		}
		return fmt.Errorf("taskkill: %s", msg)
	}
	return nil
}

func (k *windowsKiller) WaitForExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !k.IsRunning(pid) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func (k *windowsKiller) IsRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds. Check with Signal(0) equivalent.
	err = proc.Signal(os.Signal(nil))
	// If error is nil or "not supported", the process might exist.
	// Use tasklist for reliable check.
	out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
	if err != nil {
		return false
	}
	return !containsAny(string(out), "no tasks")
}
