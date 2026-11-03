//go:build !windows

package killer

import (
	"errors"
	"os/exec"
	"testing"
)

// The process listening on a port can exit between the lookup and the
// kill. That must be reported as "already gone", not as a failure.
func TestKillExitedProcessIsReportedAsGone(t *testing.T) {
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	pid := cmd.Process.Pid // exited and reaped

	err := New().Kill(pid, false)
	if err == nil {
		t.Fatalf("Kill(%d) of an exited process returned nil", pid)
	}
	if !IsProcessGone(err) || !errors.Is(err, ErrNotFound) {
		t.Fatalf("Kill(%d) = %q, want a process-gone error wrapping ErrNotFound", pid, err)
	}
}
