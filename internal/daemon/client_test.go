package daemon

import (
	"net"
	"testing"
)

func TestSelectDaemonPortUsesPreferredWhenAvailable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	preferred := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	port, err := selectDaemonPort(preferred)
	if err != nil {
		t.Fatalf("selectDaemonPort() error = %v", err)
	}
	if port != preferred {
		t.Fatalf("selectDaemonPort() = %d, want %d", port, preferred)
	}
}

func TestSelectDaemonPortFallsBackWhenPreferredIsOccupied(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer ln.Close()

	occupied := ln.Addr().(*net.TCPAddr).Port
	port, err := selectDaemonPort(occupied)
	if err != nil {
		t.Fatalf("selectDaemonPort() error = %v", err)
	}
	if port == occupied {
		t.Fatalf("selectDaemonPort() = %d, want fallback port", port)
	}
}
