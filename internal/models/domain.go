package models

import "time"

type Domain struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateDomainInput struct {
	Name string `json:"name"`
}
