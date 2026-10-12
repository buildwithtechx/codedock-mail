package imap

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

func GetQuota(ctx context.Context, addr, username, password string) (int64, error) {
	dialer := net.Dialer{Timeout: 15 * time.Second}
	conn, err := tls.DialWithDialer(&dialer, "tcp", addr, &tls.Config{
		ServerName: serverName(addr),
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return 0, fmt.Errorf("dial imap: %w", err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}
	return quotaOverConn(conn, username, password)
}

func serverName(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}

func quotaOverConn(conn net.Conn, username, password string) (int64, error) {
	reader := textproto.NewReader(bufio.NewReader(conn))
	greeting, err := reader.ReadLine()
	if err != nil {
		return 0, fmt.Errorf("read greeting: %w", err)
	}
	if !strings.HasPrefix(greeting, "*") {
		return 0, fmt.Errorf("unexpected greeting: %s", greeting)
	}
	if err := login(conn, reader, username, password); err != nil {
		return 0, err
	}
	used, err := fetchQuota(conn, reader)
	fmt.Fprintf(conn, "a003 LOGOUT\r\n")
	if err != nil {
		return 0, err
	}
	return used, nil
}

func login(conn net.Conn, reader *textproto.Reader, username, password string) error {
	fmt.Fprintf(conn, "a001 LOGIN {%d}\r\n", len(username))
	if err := expectContinuation(reader); err != nil {
		return err
	}
	fmt.Fprintf(conn, "%s {%d}\r\n", username, len(password))
	if err := expectContinuation(reader); err != nil {
		return err
	}
	fmt.Fprintf(conn, "%s\r\n", password)
	status, err := readTagged(reader, "a001")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(status, "a001 OK") {
		return fmt.Errorf("imap login rejected")
	}
	return nil
}

func fetchQuota(conn net.Conn, reader *textproto.Reader) (int64, error) {
	fmt.Fprintf(conn, "a002 GETQUOTA \"\"\r\n")
	lines, status, err := readUntilTagged(reader, "a002")
	if err != nil {
		return 0, err
	}
	if !strings.HasPrefix(status, "a002 OK") {
		return 0, fmt.Errorf("imap quota unavailable")
	}
	for _, line := range lines {
		if used, _, ok := parseQuotaLine(line); ok {
			return used, nil
		}
	}
	return 0, fmt.Errorf("imap quota response missing STORAGE")
}

func expectContinuation(reader *textproto.Reader) error {
	line, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("read continuation: %w", err)
	}
	if !strings.HasPrefix(line, "+") {
		return fmt.Errorf("expected continuation, got: %s", line)
	}
	return nil
}

func readTagged(reader *textproto.Reader, tag string) (string, error) {
	_, status, err := readUntilTagged(reader, tag)
	return status, err
}

func readUntilTagged(reader *textproto.Reader, tag string) ([]string, string, error) {
	lines := []string{}
	for {
		line, err := reader.ReadLine()
		if err != nil {
			return nil, "", fmt.Errorf("read response: %w", err)
		}
		if strings.HasPrefix(line, tag+" ") {
			return lines, line, nil
		}
		lines = append(lines, line)
	}
}

func parseQuotaLine(line string) (int64, int64, bool) {
	upper := strings.ToUpper(line)
	at := strings.Index(upper, "STORAGE")
	if at < 0 {
		return 0, 0, false
	}
	fields := strings.Fields(line[at:])
	if len(fields) < 3 {
		return 0, 0, false
	}
	used, err := strconv.ParseInt(strings.Trim(fields[1], "()"), 10, 64)
	if err != nil {
		return 0, 0, false
	}
	max, err := strconv.ParseInt(strings.Trim(fields[2], "()"), 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return used * 1024, max * 1024, true
}
