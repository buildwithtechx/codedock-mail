package models

import "time"

type DKIMKey struct {
	ID         string    `json:"id"`
	DomainID   string    `json:"domainId"`
	Selector   string    `json:"selector"`
	Algorithm  string    `json:"algorithm"`
	KeyID      string    `json:"keyId,omitempty"`
	PublicKey  string    `json:"publicKey,omitempty"`
	DNSRecord  string    `json:"dnsRecord,omitempty"`
	PrivateKey string    `json:"-"`
	CreatedAt  time.Time `json:"createdAt"`
}

type CreateDKIMInput struct {
	Selector  string `json:"selector"`
	Algorithm string `json:"algorithm"`
}
