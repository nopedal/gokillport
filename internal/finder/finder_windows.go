package finder

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type windowsFinder struct{}

func newPlatformFinder() PortFinder {
	return &windowsFinder{}
}

func (f *windowsFinder) FindByPort(port int) (*ProcessInfo, error) {
	out, err := exec.Command("netstat", "-ano", "-p", "TCP").Output()
	if err != nil {
		return nil, fmt.Errorf("netstat: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	suffix := fmt.Sprintf(":%d", port)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if !strings.HasSuffix(fields[1], suffix) {
			continue
		}
		if fields[3] != "LISTENING" {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err != nil {
			continue
		}
		name := getWindowsProcessName(pid)
		user := getWindowsProcessUser(pid)
		return &ProcessInfo{
			Port:    port,
			PID:     pid,
			Process: name,
			User:    user,
		}, nil
	}
	return nil, nil
}

func (f *windowsFinder) ListAll() ([]ProcessInfo, error) {
	out, err := exec.Command("netstat", "-ano", "-p", "TCP").Output()
	if err != nil {
		return nil, fmt.Errorf("netstat: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	var results []ProcessInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) < 5 || fields[3] != "LISTENING" {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err != nil {
			continue
		}
		// Parse port from local address
		localAddr := fields[1]
		port := 0
		if idx := strings.LastIndex(localAddr, ":"); idx != -1 {
			port, _ = strconv.Atoi(localAddr[idx+1:])
		}
		results = append(results, ProcessInfo{
			Port:    port,
			PID:     pid,
			Process: getWindowsProcessName(pid),
			User:    getWindowsProcessUser(pid),
		})
	}
	return results, nil
}

func getWindowsProcessName(pid int) string {
	out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH").Output()
	if err != nil {
		return "unknown"
	}
	line := strings.TrimSpace(string(out))
	if line == "" || strings.Contains(line, "No tasks") {
		return "unknown"
	}
	// CSV format: "name.exe","PID","Session Name","Session#","Mem Usage"
	parts := strings.Split(line, ",")
	if len(parts) > 0 {
		return strings.Trim(parts[0], "\"")
	}
	return "unknown"
}

func getWindowsProcessUser(pid int) string {
	out, err := exec.Command("wmic", "process", "where",
		fmt.Sprintf("ProcessId=%d", pid), "get", "GetOwner", "/FORMAT:VALUE").Output()
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "User=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}
