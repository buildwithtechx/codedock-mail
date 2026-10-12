package provisioner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"
)

func CheckPort25Egress(ctx context.Context) error {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", "aspmx.l.google.com:25")
	if err != nil {
		return fmt.Errorf("port 25 egress: %w", err)
	}
	conn.Close()
	return nil
}

func LookupMXHosts(ctx context.Context, domain string) ([]string, error) {
	records, err := net.DefaultResolver.LookupMX(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("lookup MX: %w", err)
	}
	hosts := make([]string, 0, len(records))
	for _, record := range records {
		hosts = append(hosts, record.Host)
	}
	return hosts, nil
}

func TCPPortClosed(ctx context.Context, addr string) bool {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return dialRefused(err)
	}
	conn.Close()
	return false
}

func dialRefused(err error) bool {
	var opErr *net.OpError
	if !errors.As(err, &opErr) || opErr.Op != "dial" {
		return false
	}
	var syscallErr *os.SyscallError
	return errors.As(opErr.Err, &syscallErr) && errors.Is(syscallErr.Err, syscall.ECONNREFUSED)
}

func CheckPTR(ctx context.Context, ip string) (string, error) {
	names, err := net.DefaultResolver.LookupAddr(ctx, ip)
	if err != nil {
		return "", fmt.Errorf("lookup PTR: %w", err)
	}
	if len(names) == 0 {
		return "", fmt.Errorf("no PTR record for %s", ip)
	}
	return names[0], nil
}
