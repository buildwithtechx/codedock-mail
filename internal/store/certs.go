package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

func (s *Store) UpsertCert(ctx context.Context, hostname string, expiresAt time.Time) (models.Cert, error) {
	var cert models.Cert
	err := s.pool.QueryRow(ctx,
		`INSERT INTO certs (id, hostname, expires_at) VALUES ($1, $2, $3)
		 ON CONFLICT (hostname) DO UPDATE SET expires_at = EXCLUDED.expires_at, updated_at = CURRENT_TIMESTAMP
		 RETURNING id, hostname, expires_at, created_at, updated_at`,
		uuid.NewString(), hostname, expiresAt.UTC()).
		Scan(&cert.ID, &cert.Hostname, &cert.ExpiresAt, &cert.CreatedAt, &cert.UpdatedAt)
	if err != nil {
		return models.Cert{}, fmt.Errorf("upsert cert: %w", err)
	}
	return cert, nil
}

func (s *Store) ListCerts(ctx context.Context) ([]models.Cert, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, hostname, expires_at, created_at, updated_at FROM certs ORDER BY expires_at`)
	if err != nil {
		return nil, fmt.Errorf("list certs: %w", err)
	}
	defer rows.Close()
	certs := []models.Cert{}
	for rows.Next() {
		var cert models.Cert
		if err := rows.Scan(&cert.ID, &cert.Hostname, &cert.ExpiresAt, &cert.CreatedAt, &cert.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan cert: %w", err)
		}
		certs = append(certs, cert)
	}
	return certs, rows.Err()
}

func (s *Store) ListExpiring(ctx context.Context, before time.Time) ([]models.Cert, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, hostname, expires_at, created_at, updated_at FROM certs WHERE expires_at < $1 ORDER BY expires_at`,
		before.UTC())
	if err != nil {
		return nil, fmt.Errorf("list expiring certs: %w", err)
	}
	defer rows.Close()
	certs := []models.Cert{}
	for rows.Next() {
		var cert models.Cert
		if err := rows.Scan(&cert.ID, &cert.Hostname, &cert.ExpiresAt, &cert.CreatedAt, &cert.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan cert: %w", err)
		}
		certs = append(certs, cert)
	}
	return certs, rows.Err()
}
