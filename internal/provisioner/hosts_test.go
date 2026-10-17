package provisioner

import (
	"strings"
	"testing"
)

func TestParseContainerHosts(t *testing.T) {
	t.Run("parses valid host-gateway, IPv4 and IPv6", func(t *testing.T) {
		input := []string{
			"mail.example.com:host-gateway",
			"db=10.2.3.4",
			"ipv6=[::1]",
		}
		res, err := ParseContainerHosts(input, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res["mail.example.com"] != "host-gateway" {
			t.Errorf("expected host-gateway, got %s", res["mail.example.com"])
		}
		if res["db"] != "10.2.3.4" {
			t.Errorf("expected 10.2.3.4, got %s", res["db"])
		}
		if res["ipv6"] != "::1" {
			t.Errorf("expected ::1, got %s", res["ipv6"])
		}
	})

	t.Run("rejects shared container network namespace", func(t *testing.T) {
		_, err := ParseContainerHosts([]string{"mail:host-gateway"}, "container:other")
		if err == nil || !strings.Contains(err.Error(), "shared container network namespace") {
			t.Fatalf("expected shared network error, got %v", err)
		}
	})

	t.Run("rejects invalid addresses and conflicting entries", func(t *testing.T) {
		_, err := ParseContainerHosts([]string{"mail:invalid_address"}, "")
		if err == nil {
			t.Fatal("expected error for invalid address")
		}

		_, err = ParseContainerHosts([]string{"mail:10.0.0.1", "mail:10.0.0.2"}, "")
		if err == nil || !strings.Contains(err.Error(), "conflicting") {
			t.Fatalf("expected conflicting error, got %v", err)
		}
	})
}

func TestConfigureExternalWebmailInstall(t *testing.T) {
	backend := ExternalWebmailBackend{
		IMAPHost: "mail.example.com",
		IMAPPort: 993,
		SMTPHost: "mail.example.com",
		SMTPPort: 587,
		UseTLS:   true,
		IsLocal:  true,
	}

	hosts, err := ConfigureExternalWebmailInstall(backend)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hosts["mail.example.com"] != "host-gateway" {
		t.Errorf("expected mail.example.com -> host-gateway, got %v", hosts["mail.example.com"])
	}
}
