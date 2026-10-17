package provisioner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureFail2banSidecar(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fail2ban-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an existing pgpass in root/.pgpass to ensure it is not touched
	existingPgpassDir := filepath.Join(tempDir, "root")
	if err := os.MkdirAll(existingPgpassDir, 0o755); err != nil {
		t.Fatalf("failed to create root dir: %v", err)
	}
	existingPgpassFile := filepath.Join(existingPgpassDir, ".pgpass")
	originalContent := "*:*:*:postgres:retained-master-pass\n"
	if err := os.WriteFile(existingPgpassFile, []byte(originalContent), 0o600); err != nil {
		t.Fatalf("failed to write existing pgpass: %v", err)
	}

	passwordWithDelimiters := `secret:pass\word`
	err = ConfigureFail2banSidecar(tempDir, "127.0.0.1", "5432", "mail", "fail2ban", passwordWithDelimiters)
	if err != nil {
		t.Fatalf("unexpected error configuring fail2ban sidecar: %v", err)
	}

	pgpassPath := filepath.Join(tempDir, "etc", "fail2ban", "codedock-postgresql.pgpass")
	confPath := filepath.Join(tempDir, "etc", "fail2ban", "codedock-postgresql.conf")

	pgpassBytes, err := os.ReadFile(pgpassPath)
	if err != nil {
		t.Fatalf("failed to read pgpass: %v", err)
	}
	expectedPgpass := "127.0.0.1:5432:mail:fail2ban:secret\\:pass\\\\word\n"
	if string(pgpassBytes) != expectedPgpass {
		t.Errorf("expected pgpass %q, got %q", expectedPgpass, string(pgpassBytes))
	}

	confBytes, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("failed to read conf: %v", err)
	}
	if !strings.Contains(string(confBytes), "export PGPORT=5432") {
		t.Errorf("expected conf to contain port 5432, got: %s", string(confBytes))
	}

	// Verify permissions
	for _, p := range []string{pgpassPath, confPath} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("failed to stat %s: %v", p, err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("expected permission 0600, got %o for %s", perm, p)
		}
	}

	// Verify existing pgpass was preserved untouched
	retained, err := os.ReadFile(existingPgpassFile)
	if err != nil {
		t.Fatalf("failed to read existing pgpass: %v", err)
	}
	if string(retained) != originalContent {
		t.Errorf("expected original pgpass retained, got %s", string(retained))
	}
}
