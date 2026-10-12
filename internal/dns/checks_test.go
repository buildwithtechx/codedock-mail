package dns

import (
	"context"
	"os"
	"testing"
)

func TestVerifyDomainLive(t *testing.T) {
	if os.Getenv("MAIL_LIVE_CHECKS") == "" {
		t.Skip("MAIL_LIVE_CHECKS is not set")
	}
	report := VerifyDomain(context.Background(), "gmail.com", "google")
	for _, check := range report.Checks {
		if check.Name == "mx" && !check.Passed {
			t.Fatalf("expected gmail MX to pass: %+v", check)
		}
	}
}

func TestMergePrefersEngine(t *testing.T) {
	local := append(ForDomain("example.com", "mail.example.com"),
		DKIMRecord("mail", "v=DKIM1; p=local"))
	engine := []Record{
		{Type: "TXT", Name: "_dmarc.example.com.", Value: "v=DMARC1; p=reject", Source: "engine"},
		{Type: "TXT", Name: "mail._domainkey.example.com.", Value: "v=DKIM1; p=engine", Source: "engine"},
	}
	merged := Merge("example.com", local, engine)
	byKey := map[string]Record{}
	for _, record := range merged {
		byKey[record.Type+" "+record.Name] = record
	}
	if got := byKey["TXT _dmarc.example.com."].Value; got != "v=DMARC1; p=reject" {
		t.Fatalf("expected engine dmarc to win, got %q", got)
	}
	if got := byKey["TXT mail._domainkey.example.com."].Value; got != "v=DKIM1; p=engine" {
		t.Fatalf("expected engine dkim to win, got %q", got)
	}
	if _, ok := byKey["MX @"]; !ok {
		t.Fatalf("expected local MX to survive, got %+v", merged)
	}
}
