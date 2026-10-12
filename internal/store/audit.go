package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

func (s *Store) Log(ctx context.Context, entry models.AuditEntry) (models.AuditEntry, error) {
	entry.ID = uuid.NewString()
	if entry.Actor == "" {
		entry.Actor = "operator"
	}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO audit_log (id, actor, action, resource_type, resource_id, detail)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`,
		entry.ID, entry.Actor, entry.Action, entry.ResourceType, entry.ResourceID, entry.Detail).
		Scan(&entry.CreatedAt)
	if err != nil {
		return models.AuditEntry{}, fmt.Errorf("insert audit entry: %w", err)
	}
	return entry, nil
}

func (s *Store) ListAudit(ctx context.Context, limit int) ([]models.AuditEntry, error) {
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, actor, action, resource_type, resource_id, detail, created_at
		 FROM audit_log ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit log: %w", err)
	}
	defer rows.Close()
	entries := []models.AuditEntry{}
	for rows.Next() {
		var entry models.AuditEntry
		if err := rows.Scan(&entry.ID, &entry.Actor, &entry.Action, &entry.ResourceType, &entry.ResourceID, &entry.Detail, &entry.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
