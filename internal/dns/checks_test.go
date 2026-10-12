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
