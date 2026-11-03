package finder

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type darwinFinder struct{}

func newPlatformFinder() PortFinder {
	return &darwinFinder{}
}

func (f *darwinFinder) FindByPort(port int) (*ProcessInfo, error) {
	out, err := exec.Command("lsof", fmt.Sprintf("-iTCP:%d", port), "-sTCP:LISTEN", "-n", "-P").Output()
	if err != nil {
		// lsof exits 1 when nothing found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("lsof: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil, nil
	}

	return parseLsofLine(lines[1], port)
}

func (f *darwinFinder) ListAll() ([]ProcessInfo, error) {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P").Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("lsof: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil, nil
	}

	seen := make(map[int]bool)
	var results []ProcessInfo
	for _, line := range lines[1:] {
		info, err := parseLsofLine(line, 0)
		if err != nil || info == nil {
			continue
		}
		if seen[info.PID] && info.Port == 0 {
			continue
		}
		seen[info.PID] = true
		results = append(results, *info)
	}
	return results, nil
}

func parseLsofLine(line string, expectedPort int) (*ProcessInfo, error) {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return nil, fmt.Errorf("unexpected lsof output: %s", line)
	}

	name := fields[0]
	pid, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid pid %q: %w", fields[1], err)
	}
	user := fields[2]

	// Parse port from the name field (e.g., *:3000 or 127.0.0.1:3000)
	port := expectedPort
	nameField := fields[8] // NAME column
	if idx := strings.LastIndex(nameField, ":"); idx != -1 {
		if p, err := strconv.Atoi(nameField[idx+1:]); err == nil {
			port = p
		}
	}

	return &ProcessInfo{
		Port:    port,
		PID:     pid,
		Process: name,
		User:    user,
	}, nil
}
