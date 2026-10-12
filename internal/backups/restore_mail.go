package backups

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrEngineRunning = errors.New("engine is running")

func RestoreMailData(dataDir, archivePath string, engineStopped func() bool) (string, error) {
	if engineStopped != nil && !engineStopped() {
		return "", fmt.Errorf("%w: stop it before restoring mail data", ErrEngineRunning)
	}
	staging, err := os.MkdirTemp(filepath.Dir(dataDir), "restore-staging-")
	if err != nil {
		return "", fmt.Errorf("create staging dir: %w", err)
	}
	defer os.RemoveAll(staging)
	if err := ExtractArchive(archivePath, staging); err != nil {
		return "", err
	}
	stamp := time.Now().UTC().Format("20060102-150405")
	restorePoint := filepath.Join(filepath.Dir(dataDir), "maildata-restore-point-"+stamp+".tar.gz")
	if _, err := os.Stat(dataDir); err == nil {
		if _, err := ArchiveDir(dataDir, restorePoint); err != nil {
			return "", fmt.Errorf("snapshot current mail data: %w", err)
		}
		if err := clearDir(dataDir); err != nil {
			return "", fmt.Errorf("clear mail data dir: %w", err)
		}
	} else {
		restorePoint = ""
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			return "", fmt.Errorf("create mail data dir: %w", err)
		}
	}
	if err := moveDirContents(staging, dataDir); err != nil {
		return restorePoint, fmt.Errorf("install restored mail data: %w", err)
	}
	return restorePoint, nil
}

func ExtractArchive(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open mail data archive: %w", err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open mail data gzip: %w", err)
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read mail data entry: %w", err)
		}
		target, err := safeJoin(destDir, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create mail data dir: %w", err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("create mail data dir: %w", err)
			}
			if err := writeFile(target, tarReader, header.FileInfo().Mode()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("refusing to extract non-regular entry: %s", header.Name)
		}
	}
}

func safeJoin(destDir, name string) (string, error) {
	flat := strings.ReplaceAll(name, "\\", "/")
	if strings.TrimSpace(flat) == "" {
		return "", fmt.Errorf("refusing to extract empty entry name")
	}
	for _, segment := range strings.Split(flat, "/") {
		if segment == ".." {
			return "", fmt.Errorf("refusing to extract escaping entry: %s", name)
		}
	}
	target := filepath.Join(destDir, filepath.Clean("/"+flat))
	if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("refusing to extract escaping entry: %s", name)
	}
	return target, nil
}

func writeFile(target string, source io.Reader, mode os.FileMode) error {
	file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
	if err != nil {
		return fmt.Errorf("write mail data file: %w", err)
	}
	_, err = io.Copy(file, source)
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("write mail data file: %w", err)
	}
	return closeErr
}

func clearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func moveDirContents(srcDir, destDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(srcDir, entry.Name()), filepath.Join(destDir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
