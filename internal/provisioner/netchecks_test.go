package provisioner

import (
	"context"
	"os"
	"testing"
)

func TestLiveNetworkChecks(t *testing.T) {
	if os.Getenv("MAIL_LIVE_CHECKS") == "" {
		t.Skip("MAIL_LIVE_CHECKS is not set")
	}
	ctx := context.Background()
	if err := CheckPort25Egress(ctx); err != nil {
		t.Fatalf("port 25 egress: %v", err)
	}
	hosts, err := LookupMXHosts(ctx, "gmail.com")
	if err != nil || len(hosts) == 0 {
		t.Fatalf("lookup MX: %v %v", hosts, err)
	}
	ptr, err := CheckPTR(ctx, "8.8.8.8")
	if err != nil || ptr == "" {
		t.Fatalf("lookup PTR: %v", err)
	}
}
