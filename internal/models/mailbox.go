package models

import "time"

type Mailbox struct {
	ID        string    `json:"id"`
	DomainID  string    `json:"domainId"`
	LocalPart string    `json:"localPart"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateMailboxInput struct {
	LocalPart string `json:"localPart"`
}

type MailboxCredentials struct {
	Mailbox
	Password string `json:"password,omitempty"`
}
