package provisioner

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type capturedRequest struct {
	method   string
	path     string
	username string
	password string
	body     map[string]any
}

func newTestServer(t *testing.T, status int, response string, captured *capturedRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()
		captured.method = r.Method
		captured.path = r.URL.EscapedPath()
		captured.username = username
		captured.password = password
		if r.Body != nil {
			raw, _ := io.ReadAll(r.Body)
			if len(raw) > 0 {
				_ = json.Unmarshal(raw, &captured.body)
			}
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
}

func TestProvisionSendsDomain(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	if err := engine.Provision("example.com"); err != nil {
		t.Fatalf("provision: %v", err)
	}
	if captured.method != http.MethodPost || captured.path != "/api/principal" {
		t.Fatalf("expected POST /api/principal, got %s %s", captured.method, captured.path)
	}
	if captured.username != "admin" || captured.password != "secret" {
		t.Fatalf("expected basic auth admin, got %s", captured.username)
	}
	if captured.body["type"] != "domain" || captured.body["name"] != "example.com" {
		t.Fatalf("unexpected body %+v", captured.body)
	}
}

func TestProvisionExistsMapsError(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{"error":"alreadyExists"}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	if err := engine.Provision("example.com"); !errors.Is(err, ErrEngineExists) {
		t.Fatalf("expected ErrEngineExists, got %v", err)
	}
}

func TestProvisionHTTPError(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 500, `boom`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	err := engine.Provision("example.com")
	if err == nil || errors.Is(err, ErrEngineExists) {
		t.Fatalf("expected plain error, got %v", err)
	}
}

func TestAddMailboxPayload(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	if err := engine.AddMailbox("alice@example.com", "$2a$10$hash", 1073741824); err != nil {
		t.Fatalf("add mailbox: %v", err)
	}
	if captured.body["type"] != "individual" || captured.body["name"] != "alice@example.com" {
		t.Fatalf("unexpected body %+v", captured.body)
	}
	if captured.body["quota"] != float64(1073741824) {
		t.Fatalf("unexpected quota %+v", captured.body["quota"])
	}
	secrets, _ := captured.body["secrets"].([]any)
	if len(secrets) != 1 || secrets[0] != "$2a$10$hash" {
		t.Fatalf("unexpected secrets %+v", captured.body["secrets"])
	}
}

func TestUpdateMailboxQuotaPayload(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	if err := engine.UpdateMailboxQuota("alice@example.com", 536870912); err != nil {
		t.Fatalf("update mailbox quota: %v", err)
	}
	if captured.method != http.MethodPatch || captured.path != "/api/principal/alice@example.com" {
		t.Fatalf("expected PATCH escaped principal path, got %s %s", captured.method, captured.path)
	}
	if captured.body["quota"] != float64(536870912) {
		t.Fatalf("unexpected quota %+v", captured.body["quota"])
	}
}

func TestRemoveMailboxPath(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	if err := engine.RemoveMailbox("alice@example.com"); err != nil {
		t.Fatalf("remove mailbox: %v", err)
	}
	if captured.method != http.MethodDelete || captured.path != "/api/principal/alice@example.com" {
		t.Fatalf("expected DELETE escaped principal path, got %s %s", captured.method, captured.path)
	}
}

func TestCreateDKIMParses(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{"data":{"selector":"mail","public_key":"abc","dns_record":"v=DKIM1; p=abc"}}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	result, err := engine.CreateDKIM("example.com", "mail", "Ed25519")
	if err != nil {
		t.Fatalf("create dkim: %v", err)
	}
	if result.Selector != "mail" || result.PublicKey != "abc" || result.DNSRecord != "v=DKIM1; p=abc" {
		t.Fatalf("unexpected result %+v", result)
	}
	if captured.body["domain"] != "example.com" || captured.body["algorithm"] != "Ed25519" {
		t.Fatalf("unexpected body %+v", captured.body)
	}
}

func TestDNSRecordsNormalizes(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{"data":[{"type":"txt","name":"_dmarc.example.com","content":"v=DMARC1"},{"type":"","name":"skip"}]}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	records, err := engine.DNSRecords("example.com")
	if err != nil {
		t.Fatalf("dns records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected one record, got %+v", records)
	}
	if records[0].Type != "TXT" || records[0].Value != "v=DMARC1" || records[0].Source != "engine" {
		t.Fatalf("unexpected record %+v", records[0])
	}
	if captured.path != "/api/dns/records/example.com" {
		t.Fatalf("unexpected path %s", captured.path)
	}
}

func TestHealth(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	if err := engine.Health(); err != nil {
		t.Fatalf("health: %v", err)
	}
	if captured.method != http.MethodGet || captured.path != "/api/principal" {
		t.Fatalf("expected GET /api/principal, got %s %s", captured.method, captured.path)
	}
}

func TestReloadCertificates(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{"data":{"warnings":{},"errors":{}}}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	if err := engine.ReloadCertificates(); err != nil {
		t.Fatalf("reload certificates: %v", err)
	}
	if captured.method != http.MethodGet || captured.path != "/api/reload/certificate" {
		t.Fatalf("expected GET /api/reload/certificate, got %s %s", captured.method, captured.path)
	}
}

func TestRemoveDomain(t *testing.T) {
	var captured capturedRequest
	server := newTestServer(t, 200, `{}`, &captured)
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	if err := engine.RemoveDomain("example.com"); err != nil {
		t.Fatalf("remove domain: %v", err)
	}
	if captured.method != http.MethodDelete || captured.path != "/api/principal/example.com" {
		t.Fatalf("expected DELETE /api/principal/example.com, got %s %s", captured.method, captured.path)
	}
}

func TestCreateDKIMFallsBackToDNSRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"data":null}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"type":"TXT","name":"mail._domainkey.example.com.","content":"v=DKIM1; k=ed25519; p=abc123"}]}`))
	}))
	defer server.Close()
	engine := NewStalwart(server.URL, "admin", "secret")
	result, err := engine.CreateDKIM("example.com", "mail", "Ed25519")
	if err != nil {
		t.Fatalf("create dkim: %v", err)
	}
	if result.DNSRecord != "v=DKIM1; k=ed25519; p=abc123" || result.PublicKey != "abc123" {
		t.Fatalf("unexpected fallback result %+v", result)
	}
}

func TestPasswordRoundTrip(t *testing.T) {
	password, err := GeneratePassword()
	if err != nil {
		t.Fatalf("generate password: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(password)
	if err != nil || len(raw) != 24 {
		t.Fatalf("expected 24 random bytes, got %q", password)
	}
	hash, err := HashSecret(password)
	if err != nil {
		t.Fatalf("hash secret: %v", err)
	}
	if !strings.HasPrefix(hash, "$2a$") {
		t.Fatalf("expected bcrypt hash, got %q", hash)
	}
}
