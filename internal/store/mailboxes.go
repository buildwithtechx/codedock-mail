package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

func (s *Store) CreateMailbox(ctx context.Context, domainID, localPart string, quotaMB int64) (models.Mailbox, error) {
	d, err := s.GetDomain(ctx, domainID)
	if err != nil {
		return models.Mailbox{}, err
	}
	m := models.Mailbox{
		ID:        uuid.NewString(),
		DomainID:  domainID,
		LocalPart: localPart,
		Address:   localPart + "@" + d.Name,
		QuotaMB:   quotaMB,
	}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO mailboxes (id, domain_id, local_part, address, quota_mb) VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`,
		m.ID, m.DomainID, m.LocalPart, m.Address, m.QuotaMB).Scan(&m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return models.Mailbox{}, ErrExists
		}
		return models.Mailbox{}, fmt.Errorf("insert mailbox: %w", err)
	}
	return m, nil
}

func (s *Store) ListMailboxes(ctx context.Context, domainID string) ([]models.Mailbox, error) {
	if _, err := s.GetDomain(ctx, domainID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, domain_id, local_part, address, quota_mb, created_at, updated_at FROM mailboxes WHERE domain_id = $1 ORDER BY address`,
		domainID)
	if err != nil {
		return nil, fmt.Errorf("list mailboxes: %w", err)
	}
	defer rows.Close()
	mailboxes := []models.Mailbox{}
	for rows.Next() {
		var m models.Mailbox
		if err := rows.Scan(&m.ID, &m.DomainID, &m.LocalPart, &m.Address, &m.QuotaMB, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan mailbox: %w", err)
		}
		mailboxes = append(mailboxes, m)
	}
	return mailboxes, rows.Err()
}

func (s *Store) GetMailbox(ctx context.Context, id string) (models.Mailbox, error) {
	var m models.Mailbox
	err := s.pool.QueryRow(ctx,
		`SELECT id, domain_id, local_part, address, quota_mb, created_at, updated_at FROM mailboxes WHERE id = $1`, id).
		Scan(&m.ID, &m.DomainID, &m.LocalPart, &m.Address, &m.QuotaMB, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Mailbox{}, ErrNotFound
		}
		return models.Mailbox{}, fmt.Errorf("get mailbox: %w", err)
	}
	return m, nil
}

func (s *Store) DeleteMailbox(ctx context.Context, id string) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM mailboxes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete mailbox: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
