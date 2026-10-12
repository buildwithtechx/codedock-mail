package certs

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"
)

func Probe(ctx context.Context, host string, port int) (time.Time, error) {
	if port < 1 || port > 65535 {
		return time.Time{}, fmt.Errorf("invalid port %d", port)
	}
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config:    &tls.Config{ServerName: host, InsecureSkipVerify: true},
	}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return time.Time{}, fmt.Errorf("tls probe %s: %w", host, err)
	}
	defer conn.Close()
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return time.Time{}, fmt.Errorf("tls probe %s: not a tls connection", host)
	}
	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return time.Time{}, fmt.Errorf("tls probe %s: no peer certificates", host)
	}
	return state.PeerCertificates[0].NotAfter, nil
}
