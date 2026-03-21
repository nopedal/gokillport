//go:build !windows

package killer

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

type unixKiller struct{}

func newPlatformKiller() ProcessKiller {
	return &unixKiller{}
}

func (k *unixKiller) Kill(pid int, force bool) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotFound, err)
	}

	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}

	if err := proc.Signal(sig); err != nil {
		if IsProcessGone(err) {
			return fmt.Errorf("%w: pid %d", ErrNotFound, pid)
		}
		if os.IsPermission(err) || IsPermissionError(err) {
			return fmt.Errorf("%w: %v", ErrPermission, err)
		}
		return err
	}
	return nil
}

func (k *unixKiller) WaitForExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !k.IsRunning(pid) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func (k *unixKiller) IsRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 checks if process exists without sending a signal
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
