package backups

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ArchiveDir(srcDir, destPath string) (int64, error) {
	info, err := os.Stat(srcDir)
	if err != nil {
		return 0, fmt.Errorf("stat mail data dir: %w", err)
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("mail data dir is not a directory: %s", srcDir)
	}
	file, err := os.Create(destPath)
	if err != nil {
		return 0, fmt.Errorf("create mail data archive: %w", err)
	}
	if err := writeArchive(file, srcDir); err != nil {
		file.Close()
		os.Remove(destPath)
		return 0, err
	}
	if err := file.Close(); err != nil {
		os.Remove(destPath)
		return 0, fmt.Errorf("close mail data archive: %w", err)
	}
	stat, err := os.Stat(destPath)
	if err != nil {
		return 0, fmt.Errorf("stat mail data archive: %w", err)
	}
	return stat.Size(), nil
}

func writeArchive(w io.Writer, srcDir string) error {
	gzipWriter := gzip.NewWriter(w)
	tarWriter := tar.NewWriter(gzipWriter)
	walkErr := filepath.WalkDir(srcDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if entry.IsDir() {
			return tarWriter.WriteHeader(dirHeader(rel))
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("refusing to archive non-regular file: %s", path)
		}
		return appendFile(tarWriter, srcDir, rel)
	})
	headerErr := tarWriter.Close()
	flushErr := gzipWriter.Close()
	if walkErr != nil {
		return fmt.Errorf("archive mail data: %w", walkErr)
	}
	if headerErr != nil {
		return fmt.Errorf("close tar writer: %w", headerErr)
	}
	if flushErr != nil {
		return fmt.Errorf("close gzip writer: %w", flushErr)
	}
	return nil
}

func dirHeader(rel string) *tar.Header {
	name := filepath.ToSlash(rel)
	if !strings.HasSuffix(name, "/") {
		name += "/"
	}
	return &tar.Header{Typeflag: tar.TypeDir, Name: name, Mode: 0o755}
}

func appendFile(tarWriter *tar.Writer, srcDir, rel string) error {
	info, err := os.Stat(filepath.Join(srcDir, rel))
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(rel)
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	file, err := os.Open(filepath.Join(srcDir, rel))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(tarWriter, file)
	return err
}
