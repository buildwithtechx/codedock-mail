package provisioner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EscapePgpassValue(val string) string {
	res := strings.ReplaceAll(val, `\`, `\\`)
	return strings.ReplaceAll(res, `:`, `\:`)
}

func ConfigureFail2banSidecar(rootDir, host, port, db, user, password string) error {
	fail2banDir := filepath.Join(rootDir, "etc", "fail2ban")
	if err := os.MkdirAll(fail2banDir, 0o755); err != nil {
		return fmt.Errorf("create fail2ban directory: %w", err)
	}

	pgpassPath := filepath.Join(fail2banDir, "codedock-postgresql.pgpass")
	confPath := filepath.Join(fail2banDir, "codedock-postgresql.conf")

	// Escape password delimiters
	escapedPassword := EscapePgpassValue(password)
	pgpassContent := fmt.Sprintf("%s:%s:%s:%s:%s\n", host, port, db, user, escapedPassword)

	if err := os.WriteFile(pgpassPath, []byte(pgpassContent), 0o600); err != nil {
		return fmt.Errorf("write fail2ban pgpass: %w", err)
	}

	confContent := fmt.Sprintf("export PGHOST=%s\nexport PGPORT=%s\nexport PGDATABASE=%s\nexport PGUSER=%s\nexport PGPASSFILE=/etc/fail2ban/codedock-postgresql.pgpass\n",
		host, port, db, user,
	)

	if err := os.WriteFile(confPath, []byte(confContent), 0o600); err != nil {
		return fmt.Errorf("write fail2ban conf: %w", err)
	}

	// Verify permissions are strictly 0600
	for _, p := range []string{pgpassPath, confPath} {
		if err := os.Chmod(p, 0o600); err != nil {
			return fmt.Errorf("chmod %s: %w", p, err)
		}
	}

	return nil
}
