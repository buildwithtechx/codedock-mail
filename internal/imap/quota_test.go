package imap

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestQuotaOverConn(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	go serveScripted(t, server, []scriptStep{
		{expectPrefix: "a001 LOGIN", send: []string{"+ ready for username"}},
		{expectPrefix: "alice", send: []string{"+ ready for password"}},
		{expectPrefix: "s3cret", send: []string{"a001 OK logged in"}},
		{expectPrefix: "a002 GETQUOTA", send: []string{`* QUOTA "" (STORAGE 512 1024)`, "a002 OK done"}},
		{expectPrefix: "a003 LOGOUT", send: []string{"a003 OK bye"}},
	})
	used, err := quotaOverConn(client, "alice", "s3cret")
	if err != nil {
		t.Fatalf("quota: %v", err)
	}
	if used != 512*1024 {
		t.Fatalf("expected %d bytes, got %d", 512*1024, used)
	}
}

func TestQuotaLoginRejected(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	go serveScripted(t, server, []scriptStep{
		{expectPrefix: "a001 LOGIN", send: []string{"+ ready for username"}},
		{expectPrefix: "alice", send: []string{"+ ready for password"}},
		{expectPrefix: "wrong", send: []string{"a001 NO bad password"}},
	})
	if _, err := quotaOverConn(client, "alice", "wrong"); err == nil {
		t.Fatalf("expected login error")
	}
}

func TestQuotaUnsupported(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	go serveScripted(t, server, []scriptStep{
		{expectPrefix: "a001 LOGIN", send: []string{"+ ready for username"}},
		{expectPrefix: "alice", send: []string{"+ ready for password"}},
		{expectPrefix: "s3cret", send: []string{"a001 OK logged in"}},
		{expectPrefix: "a002 GETQUOTA", send: []string{"a002 NO quota not supported"}},
		{expectPrefix: "a003 LOGOUT", send: []string{"a003 OK bye"}},
	})
	if _, err := quotaOverConn(client, "alice", "s3cret"); err == nil {
		t.Fatalf("expected quota error")
	}
}

func TestParseQuotaLine(t *testing.T) {
	used, max, ok := parseQuotaLine(`* QUOTA "" (STORAGE 10 100)`)
	if !ok || used != 10*1024 || max != 100*1024 {
		t.Fatalf("expected 10/100 KiB, got %d %d %v", used, max, ok)
	}
	if _, _, ok := parseQuotaLine("* CAPABILITY IMAP4rev1"); ok {
		t.Fatalf("expected no match for capability line")
	}
	if _, _, ok := parseQuotaLine(`* QUOTA "" (STORAGE broken)`); ok {
		t.Fatalf("expected no match for malformed line")
	}
}

type scriptStep struct {
	expectPrefix string
	send         []string
}

func serveScripted(t *testing.T, conn net.Conn, steps []scriptStep) {
	t.Helper()
	defer conn.Close()
	fmt.Fprintf(conn, "* OK fake imap\r\n")
	reader := bufio.NewReader(conn)
	for _, step := range steps {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		if !strings.HasPrefix(strings.TrimSpace(line), step.expectPrefix) {
			t.Errorf("expected %q, got %q", step.expectPrefix, strings.TrimSpace(line))
			return
		}
		for _, response := range step.send {
			fmt.Fprintf(conn, "%s\r\n", response)
		}
	}
}
