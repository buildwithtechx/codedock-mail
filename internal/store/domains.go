package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

func (s *Store) CreateDomain(ctx context.Context, name string) (models.Domain, error) {
	d := models.Domain{ID: uuid.NewString(), Name: name}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO domains (id, name) VALUES ($1, $2) RETURNING created_at, updated_at`,
		d.ID, d.Name).Scan(&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return models.Domain{}, ErrExists
		}
		return models.Domain{}, fmt.Errorf("insert domain: %w", err)
	}
	return d, nil
}

func (s *Store) GetDomain(ctx context.Context, id string) (models.Domain, error) {
	var d models.Domain
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at FROM domains WHERE id = $1`, id).
		Scan(&d.ID, &d.Name, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Domain{}, ErrNotFound
		}
		return models.Domain{}, fmt.Errorf("get domain: %w", err)
	}
	return d, nil
}

func (s *Store) GetDomainByName(ctx context.Context, name string) (models.Domain, error) {
	var d models.Domain
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at FROM domains WHERE name = $1`, name).
		Scan(&d.ID, &d.Name, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Domain{}, ErrNotFound
		}
		return models.Domain{}, fmt.Errorf("get domain by name: %w", err)
	}
	return d, nil
}

func (s *Store) ListDomains(ctx context.Context) ([]models.Domain, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, created_at, updated_at FROM domains ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	defer rows.Close()
	domains := []models.Domain{}
	for rows.Next() {
		var d models.Domain
		if err := rows.Scan(&d.ID, &d.Name, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan domain: %w", err)
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (s *Store) DeleteDomain(ctx context.Context, id string) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM domains WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete domain: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
