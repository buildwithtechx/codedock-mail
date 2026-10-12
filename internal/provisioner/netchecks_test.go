package provisioner

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"
	"testing"
)

func TestTCPPortClosedOnRefused(t *testing.T) {
	if !TCPPortClosed(context.Background(), "127.0.0.1:1") {
		t.Fatalf("expected refused port to read closed")
	}
}

func TestDialRefusedClassifies(t *testing.T) {
	refused := &net.OpError{Op: "dial", Err: &os.SyscallError{Syscall: "connect", Err: syscall.ECONNREFUSED}}
	if !dialRefused(refused) {
		t.Fatalf("expected refused dial to read closed")
	}
	timedOut := &net.OpError{Op: "dial", Err: &os.SyscallError{Syscall: "connect", Err: syscall.ETIMEDOUT}}
	if dialRefused(timedOut) {
		t.Fatalf("expected timeout to read unknown, not closed")
	}
	if dialRefused(&net.DNSError{Err: "no such host"}) {
		t.Fatalf("expected dns error to read unknown, not closed")
	}
	if dialRefused(errors.New("boom")) {
		t.Fatalf("expected plain error to read unknown, not closed")
	}
}

func TestTCPPortClosedOnUnresolvable(t *testing.T) {
	if TCPPortClosed(context.Background(), "invalid.invalid:993") {
		t.Fatalf("expected unresolvable host to read unknown, not closed")
	}
}

func TestLiveNetworkChecks(t *testing.T) {
	if os.Getenv("MAIL_LIVE_CHECKS") == "" {
		t.Skip("MAIL_LIVE_CHECKS is not set")
	}
	ctx := context.Background()
	if err := CheckPort25Egress(ctx); err != nil {
		t.Fatalf("port 25 egress: %v", err)
	}
	hosts, err := LookupMXHosts(ctx, "gmail.com")
	if err != nil || len(hosts) == 0 {
		t.Fatalf("lookup MX: %v %v", hosts, err)
	}
	ptr, err := CheckPTR(ctx, "8.8.8.8")
	if err != nil || ptr == "" {
		t.Fatalf("lookup PTR: %v", err)
	}
}
