package imap

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"time"
)

type IdleEvent struct {
	Folder string `json:"folder"`
	Type   string `json:"type"`
	Raw    string `json:"raw"`
}

func QuoteMailbox(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return `"INBOX"`
	}
	escaped := strings.ReplaceAll(name, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return fmt.Sprintf(`"%s"`, escaped)
}

func NormalizeFolderSlug(viewOrFolder string) string {
	slug := strings.ToLower(strings.TrimSpace(viewOrFolder))
	switch slug {
	case "inbox", "":
		return "INBOX"
	case "sent", "sent-mail":
		return "Sent"
	case "trash", "bin":
		return "Trash"
	case "drafts":
		return "Drafts"
	case "spam", "junk":
		return "Junk"
	case "archive":
		return "Archive"
	default:
		return viewOrFolder
	}
}

func IdleMailbox(
	ctx context.Context,
	conn net.Conn,
	folder string,
	onEvent func(event IdleEvent),
) error {
	folder = NormalizeFolderSlug(folder)
	reader := textproto.NewReader(bufio.NewReader(conn))

	// Select folder
	selectTag := "s001"
	if _, err := fmt.Fprintf(conn, "%s SELECT %s\r\n", selectTag, QuoteMailbox(folder)); err != nil {
		return fmt.Errorf("select mailbox: %w", err)
	}
	if _, err := readTagged(reader, selectTag); err != nil {
		return fmt.Errorf("select mailbox response: %w", err)
	}

	// Begin IDLE
	idleTag := "i001"
	if _, err := fmt.Fprintf(conn, "%s IDLE\r\n", idleTag); err != nil {
		return fmt.Errorf("send idle: %w", err)
	}
	if err := expectContinuation(reader); err != nil {
		return fmt.Errorf("idle continuation: %w", err)
	}

	doneChan := make(chan struct{})
	errChan := make(chan error, 1)

	go func() {
		defer close(doneChan)
		for {
			line, err := reader.ReadLine()
			if err != nil {
				select {
				case <-ctx.Done():
				default:
					errChan <- err
				}
				return
			}
			if strings.HasPrefix(line, idleTag+" ") {
				return
			}
			if strings.HasPrefix(line, "*") {
				eventType := classifyImapUntagged(line)
				if onEvent != nil {
					onEvent(IdleEvent{
						Folder: folder,
						Type:   eventType,
						Raw:    line,
					})
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
		_, _ = fmt.Fprintf(conn, "DONE\r\n")
		select {
		case <-doneChan:
		case <-time.After(2 * time.Second):
			_ = conn.Close()
		}
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func classifyImapUntagged(line string) string {
	upper := strings.ToUpper(line)
	switch {
	case strings.Contains(upper, "EXISTS"):
		return "exists"
	case strings.Contains(upper, "EXPUNGE"):
		return "expunge"
	case strings.Contains(upper, "FETCH"):
		return "fetch"
	case strings.Contains(upper, "RECENT"):
		return "recent"
	default:
		return "update"
	}
}
