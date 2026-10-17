package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v4"
)

func TestHandleStaticFS(t *testing.T) {
	// Setup mock FS
	mockFS := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{
			Data: []byte("<html><body>CodeDock Admin</body></html>"),
		},
		"dist/assets/main-abc12345.js": &fstest.MapFile{
			Data: []byte("console.log('main');"),
		},
		"dist/assets/main-abc12345.js.gz": &fstest.MapFile{
			Data: []byte("mock-gzipped-bytes"),
		},
	}
	DistFS = mockFS

	e := echo.New()
	RegisterHandlers(e)

	t.Run("serves index.html with no-cache and must-revalidate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		cc := rec.Header().Get("Cache-Control")
		if cc != "no-cache, must-revalidate" {
			t.Errorf("expected no-cache, must-revalidate, got %s", cc)
		}
		if rec.Header().Get("Vary") != "Accept-Encoding" {
			t.Errorf("expected Vary: Accept-Encoding, got %s", rec.Header().Get("Vary"))
		}
	})

	t.Run("serves fingerprinted asset with immutable cache control", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/assets/main-abc12345.js", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		cc := rec.Header().Get("Cache-Control")
		if cc != "public, max-age=31536000, immutable" {
			t.Errorf("expected immutable cache control, got %s", cc)
		}
	})

	t.Run("serves precompressed gzip asset when Accept-Encoding includes gzip", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/assets/main-abc12345.js", nil)
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		ce := rec.Header().Get("Content-Encoding")
		if ce != "gzip" {
			t.Errorf("expected Content-Encoding: gzip, got %s", ce)
		}
		if rec.Body.String() != "mock-gzipped-bytes" {
			t.Errorf("expected precompressed content, got %s", rec.Body.String())
		}
	})
}
