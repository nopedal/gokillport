package finder

import (
	"net"
	"os"
	"strings"
	"testing"
)

func TestParseProcNetTCPTable(t *testing.T) {
	// Excerpts of /proc/net/tcp and /proc/net/tcp6: a listener on 3000
	// (0x0BB8) and an established connection.
	tcp := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:0BB8 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 11111 1 0000000000000000 100 0 0 10 0
   1: 0100007F:C350 0100007F:0BB8 01 00000000:00000000 00:00000000 00000000  1000        0 22222 1 0000000000000000 20 4 30 10 -1
`
	tcp6 := `  sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000000000000000000000000000:1F90 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 33333 1 0000000000000000 100 0 0 10 0
`
	got := parseProcNetTCPTable(strings.NewReader(tcp))
	want := []procEntry{
		{localPort: "0BB8", state: "0A", inode: "11111"},
		{localPort: "C350", state: "01", inode: "22222"},
	}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("tcp: got %+v, want %+v", got, want)
	}

	got = parseProcNetTCPTable(strings.NewReader(tcp6))
	if len(got) != 1 || got[0] != (procEntry{localPort: "1F90", state: "0A", inode: "33333"}) {
		t.Errorf("tcp6: got %+v", got)
	}
}

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
