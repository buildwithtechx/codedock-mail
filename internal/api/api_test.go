package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/buildwithtechx/codedock-mail/internal/dns"
	"github.com/buildwithtechx/codedock-mail/internal/provisioner"
	"github.com/buildwithtechx/codedock-mail/internal/store"
)

func TestHealthIsPublic(t *testing.T) {
	srv := New(nil, nil, "secret")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAPIRequiresToken(t *testing.T) {
	srv := New(nil, nil, "secret")
	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without credentials, got %d", rec.Code)
	}
}

func TestAPIRejectsWrongToken(t *testing.T) {
	srv := New(nil, nil, "secret")
	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong token, got %d", rec.Code)
	}
}

func TestHealthReportsUnconfigured(t *testing.T) {
	srv := New(nil, nil, "secret")
	body := getHealth(t, srv)
	if body["postgres"] != "unconfigured" || body["engine"] != "unconfigured" {
		t.Fatalf("expected unconfigured components, got %+v", body)
	}
}

func TestHealthDegradedOnDatabaseError(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), "postgres://127.0.0.1:1/mail?connect_timeout=1")
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()
	srv := New(store.New(pool), nil, "secret")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestHealthDegradedOnEngineError(t *testing.T) {
	srv := New(nil, &fakeEngine{healthErr: errEngineDown}, "secret")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	body := decodeHealth(t, rec)
	if body["engine"] != "error" || body["status"] != "degraded" {
		t.Fatalf("expected degraded engine, got %+v", body)
	}
}

func TestAutoconfigIsPublic(t *testing.T) {
	srv := New(nil, nil, "secret")
	req := httptest.NewRequest(http.MethodGet, "/autoconfig/mail/config-v1.1.xml", nil)
	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing email (proving the route is public), got %d", rec.Code)
	}
}

var errEngineDown = errors.New("engine down")

func getHealth(t *testing.T, srv *Server) map[string]string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)
	return decodeHealth(t, rec)
}

func decodeHealth(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	return body
}

type fakeEngine struct {
	healthErr error
}

func (f *fakeEngine) Provision(domain string) error { return nil }

func (f *fakeEngine) RemoveDomain(domain string) error { return nil }

func (f *fakeEngine) AddMailbox(address, secretHash string, quotaBytes int64) error { return nil }

func (f *fakeEngine) UpdateMailboxQuota(address string, quotaBytes int64) error { return nil }

func (f *fakeEngine) RemoveMailbox(address string) error { return nil }

func (f *fakeEngine) CreateDKIM(domain, selector, algorithm string) (provisioner.DKIMResult, error) {
	return provisioner.DKIMResult{}, nil
}

func (f *fakeEngine) DNSRecords(domain string) ([]dns.Record, error) { return nil, nil }

func (f *fakeEngine) ReloadCertificates() error { return nil }

func (f *fakeEngine) SieveAddr() string { return "" }

func (f *fakeEngine) ImapAddr() string { return "" }

func (f *fakeEngine) Health() error { return f.healthErr }
