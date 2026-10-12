package models

import "time"

type Mailbox struct {
	ID        string    `json:"id"`
	DomainID  string    `json:"domainId"`
	LocalPart string    `json:"localPart"`
	Address   string    `json:"address"`
	QuotaMB   int64     `json:"quotaMb"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateMailboxInput struct {
	LocalPart string `json:"localPart"`
	QuotaMB   int64  `json:"quotaMb"`
}

type UpdateQuotaInput struct {
	QuotaMB int64 `json:"quotaMb"`
}

type SieveInput struct {
	Password string `json:"password"`
	Name     string `json:"name"`
	Script   string `json:"script"`
}

type MailboxCredentials struct {
	Mailbox
	Password string `json:"password,omitempty"`
}
