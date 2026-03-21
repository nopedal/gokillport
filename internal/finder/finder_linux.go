package finder

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type linuxFinder struct{}

func newPlatformFinder() PortFinder {
	return &linuxFinder{}
}

func (f *linuxFinder) FindByPort(port int) (*ProcessInfo, error) {
	// Try ss first, fall back to /proc/net/tcp
	if info, err := f.findWithSS(port); err == nil {
		return info, nil
	}
	return f.findWithProc(port)
}

func (f *linuxFinder) ListAll() ([]ProcessInfo, error) {
	if results, err := f.listWithSS(); err == nil {
		return results, nil
	}
	return f.listWithProc()
}

func (f *linuxFinder) findWithSS(port int) (*ProcessInfo, error) {
	out, err := exec.Command("ss", "-tlnp", fmt.Sprintf("sport = :%d", port)).Output()
	if err != nil {
		return nil, fmt.Errorf("ss: %w", err)
	}
	return parseSSOutput(string(out), port)
}

func (f *linuxFinder) listWithSS() ([]ProcessInfo, error) {
	out, err := exec.Command("ss", "-tlnp").Output()
	if err != nil {
		return nil, fmt.Errorf("ss: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil, nil
	}

	var results []ProcessInfo
	for _, line := range lines[1:] {
		info, err := parseSSLine(line)
		if err != nil || info == nil {
			continue
		}
		results = append(results, *info)
	}
	return results, nil
}

func parseSSOutput(output string, port int) (*ProcessInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, nil
	}

	for _, line := range lines[1:] {
		info, err := parseSSLine(line)
		if err != nil || info == nil {
			continue
		}
		if info.Port == port {
			return info, nil
		}
	}
	return nil, nil
}

func parseSSLine(line string) (*ProcessInfo, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil, fmt.Errorf("unexpected ss output")
	}

	// Local address is field 3 (0-indexed), e.g. *:3000 or 0.0.0.0:3000
	localAddr := fields[3]
	port := 0
	if idx := strings.LastIndex(localAddr, ":"); idx != -1 {
		if p, err := strconv.Atoi(localAddr[idx+1:]); err == nil {
			port = p
		}
	}

	// Process info is in the last field, e.g. users:(("node",pid=1234,fd=3))
	pid := 0
	processName := ""
	for _, field := range fields {
		if strings.HasPrefix(field, "users:") {
			pid, processName = parseSSUsers(field)
			break
		}
	}

	if pid == 0 {
		return nil, nil
	}

	user := getUserForPID(pid)

	return &ProcessInfo{
		Port:    port,
		PID:     pid,
		Process: processName,
		User:    user,
	}, nil
}

func parseSSUsers(field string) (int, string) {
	// Format: users:(("name",pid=1234,fd=3))
	pid := 0
	name := ""

	if idx := strings.Index(field, "pid="); idx != -1 {
		rest := field[idx+4:]
		if end := strings.IndexAny(rest, ",)"); end != -1 {
			pid, _ = strconv.Atoi(rest[:end])
		}
	}

	if idx := strings.Index(field, "((\""); idx != -1 {
		rest := field[idx+3:]
		if end := strings.Index(rest, "\""); end != -1 {
			name = rest[:end]
		}
	}

	return pid, name
}

// /proc/net/tcp fallback

func (f *linuxFinder) findWithProc(port int) (*ProcessInfo, error) {
	entries, err := parseProcNetTCP()
	if err != nil {
		return nil, err
	}

	hexPort := fmt.Sprintf("%04X", port)
	for _, e := range entries {
		if e.localPort == hexPort && e.state == "0A" { // 0A = LISTEN
			name := getProcessName(e.inode)
			user := getUserForPID(pidForInode(e.inode))
			pid := pidForInode(e.inode)
			if pid == 0 {
				continue
			}
			return &ProcessInfo{
				Port:    port,
				PID:     pid,
				Process: name,
				User:    user,
			}, nil
		}
	}
	return nil, nil
}

func (f *linuxFinder) listWithProc() ([]ProcessInfo, error) {
	entries, err := parseProcNetTCP()
	if err != nil {
		return nil, err
	}

	var results []ProcessInfo
	for _, e := range entries {
		if e.state != "0A" { // Only LISTEN
			continue
		}
		port, _ := strconv.ParseInt(e.localPort, 16, 32)
		pid := pidForInode(e.inode)
		if pid == 0 {
			continue
		}
		results = append(results, ProcessInfo{
			Port:    int(port),
			PID:     pid,
			Process: getProcessName(e.inode),
			User:    getUserForPID(pid),
		})
	}
	return results, nil
}

type procEntry struct {
	localPort string
	state     string
	inode     string
}

func parseProcNetTCP() ([]procEntry, error) {
	f, err := os.Open("/proc/net/tcp")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []procEntry
	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		// local_address is field 1, format: IP:PORT (hex)
		localAddr := fields[1]
		parts := strings.Split(localAddr, ":")
		if len(parts) != 2 {
			continue
		}
		entries = append(entries, procEntry{
			localPort: parts[1],
			state:     fields[3],
			inode:     fields[9],
		})
	}
	return entries, nil
}

func pidForInode(inode string) int {
	dirs, _ := os.ReadDir("/proc")
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(d.Name())
		if err != nil {
			continue
		}
		fdPath := filepath.Join("/proc", d.Name(), "fd")
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdPath, fd.Name()))
			if err != nil {
				continue
			}
			if strings.Contains(link, "socket:["+inode+"]") {
				return pid
			}
		}
	}
	return 0
}

func getProcessName(inode string) string {
	pid := pidForInode(inode)
	if pid == 0 {
		return "unknown"
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

func getUserForPID(pid int) string {
	info, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	if err != nil {
		return "unknown"
	}
	// On Linux, the file owner corresponds to the process owner
	// We read /proc/PID/status for the UID
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		_ = info
		return "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return uidToUsername(fields[1])
			}
		}
	}
	return "unknown"
}

func uidToUsername(uid string) string {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return uid
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) >= 3 && parts[2] == uid {
			return parts[0]
		}
	}
	return uid
}
