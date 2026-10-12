package certs

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestProbeReadsExpiry(t *testing.T) {
	server := httptest.NewTLSServer(nil)
	defer server.Close()
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse test url: %v", err)
	}
	port, err := strconv.Atoi(target.Port())
	if err != nil {
		t.Fatalf("parse test port: %v", err)
	}
	expiry, err := Probe(context.Background(), target.Hostname(), port)
	if err != nil {
		t.Fatalf("probe: %v", err)
	}
	if time.Until(expiry) <= 0 {
		t.Fatalf("expected future expiry, got %v", expiry)
	}
}

func TestProbeInvalidPort(t *testing.T) {
	if _, err := Probe(context.Background(), "example.com", 0); err == nil {
		t.Fatalf("expected error for invalid port")
	}
}
