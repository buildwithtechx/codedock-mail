package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

func (s *Store) CreateDKIMKey(ctx context.Context, key models.DKIMKey) (models.DKIMKey, error) {
	key.ID = uuid.NewString()
	err := s.pool.QueryRow(ctx,
		`INSERT INTO dkim_keys (id, domain_id, selector, algorithm, key_id, public_key, dns_record, private_key)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at`,
		key.ID, key.DomainID, key.Selector, key.Algorithm, key.KeyID, key.PublicKey, key.DNSRecord, key.PrivateKey).
		Scan(&key.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return models.DKIMKey{}, ErrExists
		}
		return models.DKIMKey{}, fmt.Errorf("insert dkim key: %w", err)
	}
	return key, nil
}

func (s *Store) GetDKIMKey(ctx context.Context, domainID string) (models.DKIMKey, error) {
	keys, err := s.ListDKIMKeys(ctx, domainID)
	if err != nil {
		return models.DKIMKey{}, err
	}
	if len(keys) == 0 {
		return models.DKIMKey{}, ErrNotFound
	}
	return keys[0], nil
}

func (s *Store) ListDKIMKeys(ctx context.Context, domainID string) ([]models.DKIMKey, error) {
	if _, err := s.GetDomain(ctx, domainID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, domain_id, selector, algorithm, COALESCE(key_id, ''), COALESCE(public_key, ''), COALESCE(dns_record, ''), COALESCE(private_key, ''), created_at
		 FROM dkim_keys WHERE domain_id = $1 ORDER BY created_at DESC`, domainID)
	if err != nil {
		return nil, fmt.Errorf("list dkim keys: %w", err)
	}
	defer rows.Close()
	keys := []models.DKIMKey{}
	for rows.Next() {
		var key models.DKIMKey
		if err := rows.Scan(&key.ID, &key.DomainID, &key.Selector, &key.Algorithm, &key.KeyID, &key.PublicKey, &key.DNSRecord, &key.PrivateKey, &key.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan dkim key: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}
