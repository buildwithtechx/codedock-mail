package backups

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

func TestListEmptyDir(t *testing.T) {
	manifests, err := List(filepath.Join(t.TempDir(), "missing"))
	if err != nil || len(manifests) != 0 {
		t.Fatalf("expected empty list, got %+v %v", manifests, err)
	}
}

func TestListSortsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, "a", time.Now().Add(-time.Hour))
	writeManifest(t, dir, "b", time.Now())
	manifests, err := List(dir)
	if err != nil || len(manifests) != 2 {
		t.Fatalf("expected two manifests, got %+v %v", manifests, err)
	}
	if manifests[0].ID != "b" || manifests[1].ID != "a" {
		t.Fatalf("expected newest first, got %+v", manifests)
	}
}

func TestDeleteRemovesFiles(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, "a", time.Now())
	if err := Delete(dir, "a"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	manifests, err := List(dir)
	if err != nil || len(manifests) != 0 {
		t.Fatalf("expected empty after delete, got %+v %v", manifests, err)
	}
	if err := Delete(dir, "missing"); err == nil {
		t.Fatalf("expected error for missing backup")
	}
}

func writeManifest(t *testing.T, dir, id string, created time.Time) {
	t.Helper()
	manifest := models.BackupManifest{ID: id, CreatedAt: created, Path: filepath.Join(dir, id+".sql.gz")}
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, id+".sql.gz.json"), raw, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}
