package certs

import "time"

type Renewer struct {
	hostname string
}

func New(hostname string) *Renewer {
	return &Renewer{hostname: hostname}
}

func (r *Renewer) Due(expiresAt time.Time) bool {
	return time.Until(expiresAt) < 30*24*time.Hour
}
