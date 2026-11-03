package finder

import (
	"net"
	"os"
	"testing"
)

func TestFindByPortFindsTCPListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	info, err := New().FindByPort(port)
	if err != nil {
		t.Fatal(err)
	}
	if info == nil || info.PID != os.Getpid() || info.Port != port {
		t.Fatalf("FindByPort(%d) = %+v, want pid %d", port, info, os.Getpid())
	}
}

// A UDP socket bound to the port is not a listener; killing its owner would
// hit an unrelated process (e.g. anything doing mDNS on 5353).
func TestFindByPortIgnoresUDP(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	port := pc.LocalAddr().(*net.UDPAddr).Port

	info, err := New().FindByPort(port)
	if err != nil {
		t.Fatal(err)
	}
	if info != nil {
		t.Fatalf("FindByPort(%d) = %+v for a UDP-only port, want nil", port, info)
	}
}

func TestListAllExcludesUDP(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	port := pc.LocalAddr().(*net.UDPAddr).Port

	procs, err := New().ListAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range procs {
		if p.PID == os.Getpid() && p.Port == port {
			t.Fatalf("ListAll includes UDP socket on port %d: %+v", port, p)
		}
	}
}
