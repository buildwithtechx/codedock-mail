package imap

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNormalizeFolderSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "INBOX"},
		{"inbox", "INBOX"},
		{"sent", "Sent"},
		{"trash", "Trash"},
		{"bin", "Trash"},
		{"junk", "Junk"},
		{"spam", "Junk"},
		{"archive", "Archive"},
		{"CustomFolder", "CustomFolder"},
	}

	for _, tc := range tests {
		got := NormalizeFolderSlug(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeFolderSlug(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestQuoteMailbox(t *testing.T) {
	if got := QuoteMailbox(""); got != `"INBOX"` {
		t.Errorf("QuoteMailbox empty = %q; want %q", got, `"INBOX"`)
	}
	if got := QuoteMailbox(`Special "Folder"`); got != `"Special \"Folder\""` {
		t.Errorf("QuoteMailbox quotes = %q; want %q", got, `"Special \"Folder\""`)
	}
}

func TestIdleMailbox(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan IdleEvent, 5)

	go func() {
		// Mock server behavior
		buf := make([]byte, 1024)
		// Read SELECT
		n, _ := serverConn.Read(buf)
		req := string(buf[:n])
		if !strings.Contains(req, "SELECT") {
			return
		}
		serverConn.Write([]byte("s001 OK [READ-WRITE] Select completed\r\n"))

		// Read IDLE
		n, _ = serverConn.Read(buf)
		req = string(buf[:n])
		if !strings.Contains(req, "IDLE") {
			return
		}
		serverConn.Write([]byte("+ idling\r\n"))

		// Send IMAP untagged notification
		time.Sleep(10 * time.Millisecond)
		serverConn.Write([]byte("* 42 EXISTS\r\n"))
		serverConn.Write([]byte("* 2 EXPUNGE\r\n"))

		// Expect DONE when canceled
		n, _ = serverConn.Read(buf)
		if strings.Contains(string(buf[:n]), "DONE") {
			serverConn.Write([]byte("i001 OK Idle terminated\r\n"))
		}
	}()

	errChan := make(chan error, 1)
	go func() {
		errChan <- IdleMailbox(ctx, clientConn, "inbox", func(evt IdleEvent) {
			events <- evt
		})
	}()

	select {
	case evt := <-events:
		if evt.Type != "exists" || !strings.Contains(evt.Raw, "42 EXISTS") {
			t.Errorf("unexpected first event: %+v", evt)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for exists event")
	}

	select {
	case evt := <-events:
		if evt.Type != "expunge" || !strings.Contains(evt.Raw, "2 EXPUNGE") {
			t.Errorf("unexpected second event: %+v", evt)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for expunge event")
	}

	cancel()

	select {
	case err := <-errChan:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for IdleMailbox to terminate on cancellation")
	}
}
