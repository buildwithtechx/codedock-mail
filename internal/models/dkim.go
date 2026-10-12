package models

import "time"

type DKIMKey struct {
	ID         string    `json:"id"`
	DomainID   string    `json:"domainId"`
	Selector   string    `json:"selector"`
	PublicKey  string    `json:"publicKey"`
	PrivateKey string    `json:"privateKey"`
	CreatedAt  time.Time `json:"createdAt"`
}
