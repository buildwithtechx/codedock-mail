package models

import "time"

type Cert struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
