package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestAutoconfigIsPublic(t *testing.T) {
	srv := New(nil, nil, "secret")
	req := httptest.NewRequest(http.MethodGet, "/autoconfig/mail/config-v1.1.xml", nil)
	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing email (proving the route is public), got %d", rec.Code)
	}
}
