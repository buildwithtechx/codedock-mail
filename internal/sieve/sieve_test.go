package sieve

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"testing"
)

func serveScripted(t *testing.T, password string) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		write := func(line string) {
			fmt.Fprintf(conn, "%s\r\n", line)
		}
		write(`OK "FakeSieve" "SASL PLAIN"`)
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(line) != `AUTHENTICATE "PLAIN"` {
			write(`NO "expected authenticate"`)
			return
		}
		write(`+ ""`)
		response, _ := reader.ReadString('\n')
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(response))
		if err != nil || string(decoded) != "\x00alice@example.com\x00"+password {
			write(`NO "authentication failed"`)
			return
		}
		write(`OK "authenticated"`)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			switch {
			case strings.HasPrefix(line, "LISTSCRIPTS"):
				write(`"main" ACTIVE`)
				write(`"old"`)
				write(`OK "listed"`)
			case strings.HasPrefix(line, "GETSCRIPT"):
				write(`{11}`)
				fmt.Fprintf(conn, "hello sieve\r\n")
				write(`OK "fetched"`)
			case strings.HasPrefix(line, "PUTSCRIPT"):
				start := strings.Index(line, "{")
				end := strings.Index(line, "+}")
				size := 0
				fmt.Sscanf(line[start+1:end], "%d", &size)
				body := make([]byte, size)
				for read := 0; read < size; {
					n, err := reader.Read(body[read:])
					if err != nil {
						return
					}
					read += n
				}
				if string(body) != "hello" {
					write(`NO "bad content"`)
					return
				}
				write(`OK "stored"`)
			case strings.HasPrefix(line, "SETACTIVE"):
				write(`OK "active"`)
			case strings.HasPrefix(line, "LOGOUT"):
				write(`OK "bye"`)
				return
			default:
				write(`NO "unknown"`)
				return
			}
		}
	}()
	return listener.Addr().String(), func() { listener.Close() }
}

func TestGetActive(t *testing.T) {
	addr, cleanup := serveScripted(t, "secret")
	defer cleanup()
	scripts, content, err := GetActive(addr, "alice@example.com", "secret")
	if err != nil {
		t.Fatalf("get active: %v", err)
	}
	if len(scripts) != 2 || !scripts[0].Active || content != "hello sieve" {
		t.Fatalf("unexpected scripts %+v content %q", scripts, content)
	}
}

func TestGetActiveBadPassword(t *testing.T) {
	addr, cleanup := serveScripted(t, "secret")
	defer cleanup()
	if _, _, err := GetActive(addr, "alice@example.com", "wrong"); err == nil {
		t.Fatalf("expected authentication error")
	}
}

func TestSetActiveScript(t *testing.T) {
	addr, cleanup := serveScripted(t, "secret")
	defer cleanup()
	if err := SetActiveScript(addr, "alice@example.com", "secret", "new", "hello"); err != nil {
		t.Fatalf("set active script: %v", err)
	}
}
