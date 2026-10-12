package backups

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreMailDataRoundTrip(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(filepath.Join(dataDir, "db"), 0o755); err != nil {
		t.Fatalf("make data dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "db", "live.sst"), []byte("live"), 0o644); err != nil {
		t.Fatalf("write live file: %v", err)
	}
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "restored.sst"), []byte("restored"), 0o644); err != nil {
		t.Fatalf("write src file: %v", err)
	}
	archive := filepath.Join(t.TempDir(), "maildata.tar.gz")
	if _, err := ArchiveDir(src, archive); err != nil {
		t.Fatalf("archive: %v", err)
	}
	restorePoint, err := RestoreMailData(dataDir, archive, func() bool { return true })
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if restorePoint == "" {
		t.Fatalf("expected restore point path")
	}
	if _, err := os.Stat(restorePoint); err != nil {
		t.Fatalf("expected restore point file: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dataDir, "restored.sst"))
	if err != nil || string(raw) != "restored" {
		t.Fatalf("expected restored file, got %q %v", raw, err)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "db")); !os.IsNotExist(err) {
		t.Fatalf("expected old contents cleared, stat err %v", err)
	}
	verifyArchiveContains(t, restorePoint, "db/live.sst")
}

func TestRestoreMailDataRefusesRunningEngine(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("make data dir: %v", err)
	}
	if _, err := RestoreMailData(dataDir, "missing.tar.gz", func() bool { return false }); err == nil {
		t.Fatalf("expected refusal while engine runs")
	}
}

func TestExtractArchiveRejectsTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "evil.tar.gz")
	writeRawArchive(t, archive, []string{"../escape.txt", "ok.txt"})
	if err := ExtractArchive(archive, t.TempDir()); err == nil {
		t.Fatalf("expected traversal rejection")
	}
}

func TestExtractArchiveRejectsSymlink(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "link.tar.gz")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "target"}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	tarWriter.Close()
	gzipWriter.Close()
	file.Close()
	if err := ExtractArchive(archive, t.TempDir()); err == nil {
		t.Fatalf("expected symlink rejection")
	}
}

func verifyArchiveContains(t *testing.T, archivePath, name string) {
	t.Helper()
	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("open restore point: %v", err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("open gzip: %v", err)
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			t.Fatalf("entry %s missing: %v", name, err)
		}
		if header.Name == name {
			return
		}
	}
}

func writeRawArchive(t *testing.T, archivePath string, names []string) {
	t.Helper()
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, name := range names {
		content := []byte("x")
		if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}); err != nil {
			t.Fatalf("write header: %v", err)
		}
		if _, err := tarWriter.Write(content); err != nil {
			t.Fatalf("write body: %v", err)
		}
	}
	tarWriter.Close()
	gzipWriter.Close()
	file.Close()
}
