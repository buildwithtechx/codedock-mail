package backups

import (
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

var backupTables = []string{"domains", "mailboxes", "dkim_keys", "certs", "audit_log"}

type Manifest struct {
	Manifest models.BackupManifest
	Path     string
}

func Dir() string {
	if dir := os.Getenv("BACKUP_DIR"); dir != "" {
		return dir
	}
	return filepath.Join("data", "backups")
}

func Create(ctx context.Context, dsn, dir string) (models.BackupManifest, error) {
	if _, err := exec.LookPath("pg_dump"); err != nil {
		return models.BackupManifest{}, fmt.Errorf("pg_dump not found: %w", err)
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return models.BackupManifest{}, fmt.Errorf("parse database url: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return models.BackupManifest{}, fmt.Errorf("create backup dir: %w", err)
	}
	id := randomID()
	path := filepath.Join(dir, "mail-"+time.Now().UTC().Format("20060102-150405")+"-"+id+".sql.gz")
	dumpArgs := []string{
		"-h", cfg.Host, "-p", fmt.Sprint(cfg.Port), "-U", cfg.User, "-d", cfg.Database,
		"--clean", "--if-exists", "--no-owner", "--no-privileges",
	}
	for _, table := range backupTables {
		dumpArgs = append(dumpArgs, "-t", table)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	dump := exec.CommandContext(ctx, "pg_dump", dumpArgs...)
	dump.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)
	stdout, err := dump.StdoutPipe()
	if err != nil {
		return models.BackupManifest{}, fmt.Errorf("pg_dump pipe: %w", err)
	}
	if err := dump.Start(); err != nil {
		return models.BackupManifest{}, fmt.Errorf("start pg_dump: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return models.BackupManifest{}, fmt.Errorf("create backup file: %w", err)
	}
	writer := gzip.NewWriter(file)
	_, copyErr := io.Copy(writer, stdout)
	closeErr := writer.Close()
	fileCloseErr := file.Close()
	waitErr := dump.Wait()
	if copyErr != nil || closeErr != nil || fileCloseErr != nil || waitErr != nil {
		os.Remove(path)
		return models.BackupManifest{}, fmt.Errorf("pg_dump failed: copy=%v close=%v file=%v wait=%w", copyErr, closeErr, fileCloseErr, waitErr)
	}
	info, err := os.Stat(path)
	if err != nil {
		return models.BackupManifest{}, fmt.Errorf("stat backup file: %w", err)
	}
	manifest := models.BackupManifest{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		Path:      path,
		SizeBytes: info.Size(),
		Tables:    backupTables,
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return models.BackupManifest{}, fmt.Errorf("encode manifest: %w", err)
	}
	if err := os.WriteFile(path+".json", encoded, 0o644); err != nil {
		return models.BackupManifest{}, fmt.Errorf("write manifest: %w", err)
	}
	return manifest, nil
}

func List(dir string) ([]models.BackupManifest, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.BackupManifest{}, nil
		}
		return nil, fmt.Errorf("read backup dir: %w", err)
	}
	manifests := []models.BackupManifest{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var manifest models.BackupManifest
		if err := json.Unmarshal(raw, &manifest); err != nil {
			continue
		}
		manifests = append(manifests, manifest)
	}
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].CreatedAt.After(manifests[j].CreatedAt)
	})
	return manifests, nil
}

func Restore(ctx context.Context, dsn, dir, id string) error {
	if _, err := exec.LookPath("psql"); err != nil {
		return fmt.Errorf("psql not found: %w", err)
	}
	manifests, err := List(dir)
	if err != nil {
		return err
	}
	path := ""
	for _, manifest := range manifests {
		if manifest.ID == id {
			path = manifest.Path
			break
		}
	}
	if path == "" {
		return fmt.Errorf("backup %s not found", id)
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse database url: %w", err)
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open backup file: %w", err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open backup gzip: %w", err)
	}
	defer reader.Close()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	restore := exec.CommandContext(ctx, "psql",
		"-h", cfg.Host, "-p", fmt.Sprint(cfg.Port), "-U", cfg.User, "-d", cfg.Database,
		"-v", "ON_ERROR_STOP=1", "--quiet")
	restore.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)
	restore.Stdin = reader
	if out, err := restore.CombinedOutput(); err != nil {
		return fmt.Errorf("psql restore failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func Delete(dir, id string) error {
	manifests, err := List(dir)
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		if manifest.ID == id {
			if err := os.Remove(manifest.Path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove backup file: %w", err)
			}
			if err := os.Remove(manifest.Path + ".json"); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove manifest: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("backup %s not found", id)
}

func randomID() string {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Sprint(time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(raw)
}
