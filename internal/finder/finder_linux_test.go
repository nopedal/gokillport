package finder

import (
	"net"
	"os"
	"testing"
)

// Servers listening on the IPv6 wildcard or loopback only show up in
// /proc/net/tcp6, so the /proc fallback (used when ss is missing) must
// read it too.
func TestFindWithProcFindsIPv6Listener(t *testing.T) {
	ln, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skipf("no IPv6 loopback: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	info, err := (&linuxFinder{}).findWithProc(port)
	if err != nil {
		t.Fatal(err)
	}
	if info == nil || info.PID != os.Getpid() {
		t.Fatalf("findWithProc(%d) = %+v, want pid %d", port, info, os.Getpid())
	}

	procs, err := (&linuxFinder{}).listWithProc()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range procs {
		if p.Port == port && p.PID == os.Getpid() {
			return
		}
	}
	t.Fatalf("listWithProc does not include the IPv6 listener on port %d", port)
}
