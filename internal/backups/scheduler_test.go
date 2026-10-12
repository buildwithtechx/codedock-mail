package backups

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneKeepsNewest(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	writeManifest(t, dir, "old", now.Add(-3*time.Hour))
	writeManifest(t, dir, "mid", now.Add(-2*time.Hour))
	writeManifest(t, dir, "new", now.Add(-time.Hour))
	if err := Prune(dir, 2); err != nil {
		t.Fatalf("prune: %v", err)
	}
	manifests, err := List(dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(manifests) != 2 || manifests[0].ID != "new" || manifests[1].ID != "mid" {
		t.Fatalf("expected new and mid, got %+v", manifests)
	}
	if _, err := os.Stat(filepath.Join(dir, "old.sql.gz.json")); !os.IsNotExist(err) {
		t.Fatalf("expected old manifest removed, stat err %v", err)
	}
}

func TestPruneKeepsAllWhenUnderLimit(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, "a", time.Now())
	if err := Prune(dir, 7); err != nil {
		t.Fatalf("prune: %v", err)
	}
	manifests, err := List(dir)
	if err != nil || len(manifests) != 1 {
		t.Fatalf("expected one manifest, got %+v %v", manifests, err)
	}
}

func TestNewSchedulerRejectsBadSpec(t *testing.T) {
	if _, err := NewScheduler("not a schedule", 7, "", t.TempDir(), "", nil); err == nil {
		t.Fatalf("expected error for bad schedule")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	scheduler, err := NewScheduler("0 3 * * *", 7, "", t.TempDir(), "", nil)
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}
	scheduler.Start()
	scheduler.Stop()
}
