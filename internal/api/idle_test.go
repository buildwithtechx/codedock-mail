package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMailIdleSSE(t *testing.T) {
	srv := New(nil, nil, "")

	req := httptest.NewRequest(http.MethodGet, "/api/mail/idle?folder=inbox", nil)
	ctx, cancel := context.WithTimeout(req.Context(), 100*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	headers := rec.Header()
	if headers.Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %s", headers.Get("Content-Type"))
	}
	if headers.Get("X-Accel-Buffering") != "no" {
		t.Errorf("expected X-Accel-Buffering: no, got %s", headers.Get("X-Accel-Buffering"))
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: mailbox") {
		t.Errorf("expected mailbox event in body, got: %s", body)
	}
	if !strings.Contains(body, `"folder":"INBOX"`) {
		t.Errorf("expected INBOX folder in payload, got: %s", body)
	}
}
