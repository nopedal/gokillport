package finder

// ProcessInfo holds information about a process bound to a port.
type ProcessInfo struct {
	Port    int    `json:"port"`
	PID     int    `json:"pid"`
	Process string `json:"process"`
	User    string `json:"user"`
}

// PortFinder finds processes listening on ports.
type PortFinder interface {
	// FindByPort returns the process listening on the given port, or nil if none.
	FindByPort(port int) (*ProcessInfo, error)

	// ListAll returns all processes currently bound to ports.
	ListAll() ([]ProcessInfo, error)
}

// New returns a PortFinder for the current platform.
func New() PortFinder {
	return newPlatformFinder()
}
