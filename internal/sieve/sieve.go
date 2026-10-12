package sieve

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

type Script struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type Client struct {
	conn net.Conn
	text *textproto.Reader
	out  *textproto.Writer
}

func Dial(ctx context.Context, addr string) (*Client, error) {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("sieve: dial %s: %w", addr, err)
	}
	client := &Client{conn: conn}
	client.text = textproto.NewReader(bufio.NewReader(conn))
	client.out = textproto.NewWriter(bufio.NewWriter(conn))
	line, err := client.text.ReadLine()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("sieve: read greeting: %w", err)
	}
	if !strings.HasPrefix(line, "OK") {
		conn.Close()
		return nil, fmt.Errorf("sieve: bad greeting: %s", line)
	}
	return client, nil
}

func (c *Client) Close() error {
	_ = c.sendLine("LOGOUT")
	return c.conn.Close()
}

func (c *Client) Authenticate(username, password string) error {
	if err := c.sendLine(fmt.Sprintf("AUTHENTICATE %s", quote("PLAIN"))); err != nil {
		return err
	}
	challenge, err := c.text.ReadLine()
	if err != nil {
		return fmt.Errorf("sieve: read auth challenge: %w", err)
	}
	if !strings.HasPrefix(challenge, "+") {
		return fmt.Errorf("sieve: unexpected auth challenge: %s", challenge)
	}
	raw := "\x00" + username + "\x00" + password
	if err := c.sendLine(base64.StdEncoding.EncodeToString([]byte(raw))); err != nil {
		return err
	}
	return c.expectOK()
}

func (c *Client) ListScripts() ([]Script, error) {
	if err := c.sendLine("LISTSCRIPTS"); err != nil {
		return nil, err
	}
	scripts := []Script{}
	for {
		line, err := c.text.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("sieve: list scripts: %w", err)
		}
		if done, err := c.statusLine(line); done || err != nil {
			return scripts, err
		} else {
			name, active := parseListLine(line)
			scripts = append(scripts, Script{Name: name, Active: active})
		}
	}
}

func (c *Client) GetScript(name string) (string, error) {
	if err := c.sendLine(fmt.Sprintf("GETSCRIPT %s", quote(name))); err != nil {
		return "", err
	}
	line, err := c.text.ReadLine()
	if err != nil {
		return "", fmt.Errorf("sieve: get script: %w", err)
	}
	if done, err := c.statusLine(line); done || err != nil {
		return "", err
	}
	size, err := parseLiteral(line)
	if err != nil {
		return "", err
	}
	content := make([]byte, size+2)
	if _, err := io.ReadFull(c.text.R, content); err != nil {
		return "", fmt.Errorf("sieve: read script: %w", err)
	}
	status, err := c.text.ReadLine()
	if err != nil {
		return "", fmt.Errorf("sieve: read script status: %w", err)
	}
	if _, err := c.statusLine(status); err != nil {
		return "", err
	}
	return string(content[:size]), nil
}

func (c *Client) PutScript(name, content string) error {
	if len(content) > 64<<10 {
		return fmt.Errorf("sieve: script too large")
	}
	if err := c.sendLine(fmt.Sprintf("PUTSCRIPT %s {%d+}", quote(name), len(content))); err != nil {
		return err
	}
	if _, err := c.out.W.WriteString(content); err != nil {
		return fmt.Errorf("sieve: send script: %w", err)
	}
	if err := c.out.W.Flush(); err != nil {
		return fmt.Errorf("sieve: flush script: %w", err)
	}
	return c.expectOK()
}

func (c *Client) SetActive(name string) error {
	if err := c.sendLine(fmt.Sprintf("SETACTIVE %s", quote(name))); err != nil {
		return err
	}
	return c.expectOK()
}

func GetActive(addr, username, password string) ([]Script, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := Dial(ctx, addr)
	if err != nil {
		return nil, "", err
	}
	defer client.Close()
	if err := client.Authenticate(username, password); err != nil {
		return nil, "", err
	}
	scripts, err := client.ListScripts()
	if err != nil {
		return nil, "", err
	}
	content := ""
	for _, script := range scripts {
		if script.Active {
			content, err = client.GetScript(script.Name)
			if err != nil {
				return nil, "", err
			}
			break
		}
	}
	return scripts, content, nil
}

func SetActiveScript(addr, username, password, name, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := Dial(ctx, addr)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Authenticate(username, password); err != nil {
		return err
	}
	if err := client.PutScript(name, content); err != nil {
		return err
	}
	return client.SetActive(name)
}

func (c *Client) sendLine(line string) error {
	if err := c.conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return fmt.Errorf("sieve: set deadline: %w", err)
	}
	if err := c.out.PrintfLine("%s", line); err != nil {
		return fmt.Errorf("sieve: send: %w", err)
	}
	return nil
}

func (c *Client) expectOK() error {
	line, err := c.text.ReadLine()
	if err != nil {
		return fmt.Errorf("sieve: read status: %w", err)
	}
	_, err = c.statusLine(line)
	return err
}

func (c *Client) statusLine(line string) (bool, error) {
	switch {
	case strings.HasPrefix(line, "OK"):
		return true, nil
	case strings.HasPrefix(line, "NO"):
		return true, fmt.Errorf("sieve: %s", line)
	case strings.HasPrefix(line, "BYE"):
		return true, fmt.Errorf("sieve: server closed connection: %s", line)
	default:
		return false, nil
	}
}

func quote(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", `"`, `\"`)
	return `"` + replacer.Replace(value) + `"`
}

func parseListLine(line string) (string, bool) {
	name := line
	active := false
	if index := strings.LastIndex(line, " "); index > 0 {
		name = line[:index]
		active = strings.TrimSpace(line[index+1:]) == "ACTIVE"
	}
	name = strings.TrimSpace(name)
	if len(name) >= 2 && strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`) {
		unquoted, err := strconv.Unquote(name)
		if err == nil {
			name = unquoted
		}
	}
	return name, active
}

func parseLiteral(line string) (int, error) {
	start := strings.Index(line, "{")
	end := strings.Index(line, "}")
	if start < 0 || end < 0 || end <= start+1 {
		return 0, fmt.Errorf("sieve: expected literal, got %s", line)
	}
	size, err := strconv.Atoi(line[start+1 : end])
	if err != nil || size < 0 {
		return 0, fmt.Errorf("sieve: bad literal size in %s", line)
	}
	return size, nil
}
