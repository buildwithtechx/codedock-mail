package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvSetsMissingKeys(t *testing.T) {
	dir := t.TempDir()
	content := "# comment line\n\nDATABASE_URL=postgres://example/db\nEMPTY=\nQUOTED=\"hello world\"\nSINGLE='a=b'\nSPACED = padded\nNOEQUALS\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o644); err != nil {
		t.Fatalf("write dotenv: %v", err)
	}
	t.Chdir(dir)
	loadDotEnv(".env")
	cases := map[string]string{
		"DATABASE_URL": "postgres://example/db",
		"EMPTY":        "",
		"QUOTED":       "hello world",
		"SINGLE":       "a=b",
		"SPACED":       "padded",
	}
	for key, want := range cases {
		if got := os.Getenv(key); got != want {
			t.Fatalf("expected %s=%q, got %q", key, want, got)
		}
	}
	if _, exists := os.LookupEnv("NOEQUALS"); exists {
		t.Fatalf("expected NOEQUALS to be skipped")
	}
}

func TestLoadDotEnvKeepsExistingEnv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("PORT=9999\n"), 0o644); err != nil {
		t.Fatalf("write dotenv: %v", err)
	}
	t.Chdir(dir)
	t.Setenv("PORT", "8081")
	loadDotEnv(".env")
	if got := os.Getenv("PORT"); got != "8081" {
		t.Fatalf("expected explicit env to win, got %q", got)
	}
}

func TestLoadDotEnvMissingFile(t *testing.T) {
	t.Chdir(t.TempDir())
	loadDotEnv(".env")
}
