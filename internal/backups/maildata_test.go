package backups

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveDirRoundTrip(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "root.txt"), []byte("root"), 0o644); err != nil {
		t.Fatalf("write root file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(src, "nested"), 0o755); err != nil {
		t.Fatalf("make nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "nested", "leaf.txt"), []byte("leaf"), 0o644); err != nil {
		t.Fatalf("write nested file: %v", err)
	}
	dest := filepath.Join(t.TempDir(), "maildata.tar.gz")
	size, err := ArchiveDir(src, dest)
	if err != nil {
		t.Fatalf("archive dir: %v", err)
	}
	if size <= 0 {
		t.Fatalf("expected positive archive size, got %d", size)
	}
	contents := readArchive(t, dest)
	if contents["root.txt"] != "root" || contents["nested/leaf.txt"] != "leaf" {
		t.Fatalf("unexpected archive contents %+v", contents)
	}
}

func TestArchiveDirRejectsMissingDir(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "maildata.tar.gz")
	if _, err := ArchiveDir(filepath.Join(t.TempDir(), "missing"), dest); err == nil {
		t.Fatalf("expected error for missing dir")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("expected no archive file, stat err %v", err)
	}
}

func TestArchiveDirRejectsFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := ArchiveDir(src, filepath.Join(t.TempDir(), "maildata.tar.gz")); err == nil {
		t.Fatalf("expected error for non-directory source")
	}
}

func readArchive(t *testing.T, path string) map[string]string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("open gzip: %v", err)
	}
	defer reader.Close()
	contents := map[string]string{}
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return contents
		}
		if err != nil {
			t.Fatalf("read tar entry: %v", err)
		}
		if !header.FileInfo().Mode().IsRegular() {
			continue
		}
		raw, err := io.ReadAll(tarReader)
		if err != nil {
			t.Fatalf("read entry %s: %v", header.Name, err)
		}
		contents[header.Name] = string(raw)
	}
}
