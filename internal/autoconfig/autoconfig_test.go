package autoconfig

import (
	"encoding/xml"
	"strings"
	"testing"
)

func testProvider() Provider {
	return Provider{Domain: "example.com", DisplayName: "Example Mail", MailHost: "mail.example.com"}
}

func TestThunderbirdXML(t *testing.T) {
	raw, err := ThunderbirdXML(testProvider())
	if err != nil {
		t.Fatalf("thunderbird xml: %v", err)
	}
	body := string(raw)
	for _, want := range []string{"mail.example.com", "993", "587", "SSL", "STARTTLS", "%EMAILADDRESS%", "example.com"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %q in thunderbird config:\n%s", want, body)
		}
	}
	var parsed tbConfig
	if err := xml.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("thunderbird config should round-trip: %v", err)
	}
	if parsed.Provider.Incoming.Port != 993 || parsed.Provider.Outgoing.Port != 587 {
		t.Fatalf("unexpected ports %+v", parsed.Provider)
	}
}

func TestOutlookXML(t *testing.T) {
	raw, err := OutlookXML(testProvider(), "alice@example.com")
	if err != nil {
		t.Fatalf("outlook xml: %v", err)
	}
	body := string(raw)
	for _, want := range []string{"mail.example.com", "alice@example.com", "IMAP", "SMTP", "responseschema/2006a"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %q in outlook config:\n%s", want, body)
		}
	}
}

func TestOutlookRequestEmailWithNamespace(t *testing.T) {
	raw := []byte(`<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006"><Request><EMailAddress>alice@example.com</EMailAddress><AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema></Request></Autodiscover>`)
	if got := OutlookRequestEmail(raw); got != "alice@example.com" {
		t.Fatalf("expected namespaced email extraction, got %q", got)
	}
}

func TestOutlookRequestEmailInvalid(t *testing.T) {
	if got := OutlookRequestEmail([]byte("not xml")); got != "" {
		t.Fatalf("expected empty email for invalid xml, got %q", got)
	}
}

func TestDomainOf(t *testing.T) {
	domain, err := DomainOf("Alice@Example.COM")
	if err != nil || domain != "example.com" {
		t.Fatalf("expected example.com, got %q %v", domain, err)
	}
	for _, bad := range []string{"alice", "alice@", "@example.com", "alice@localhost", ""} {
		if _, err := DomainOf(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}
