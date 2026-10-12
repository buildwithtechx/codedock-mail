package provisioner

import (
	"fmt"
	"net/http"
	"time"
)

type StalwartEngine struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewStalwart(baseURL, token string) *StalwartEngine {
	return &StalwartEngine{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *StalwartEngine) Provision(domain string) error {
	return fmt.Errorf("stalwart: provision %s not implemented", domain)
}

func (e *StalwartEngine) AddMailbox(address string) error {
	return fmt.Errorf("stalwart: add mailbox %s not implemented", address)
}

func (e *StalwartEngine) Health() error {
	return fmt.Errorf("stalwart: health check not implemented")
}
