package models

import "github.com/buildwithtechx/codedock-mail/internal/dns"

type DNSResponse struct {
	Domain      string       `json:"domain"`
	Records     []dns.Record `json:"records"`
	EngineError string       `json:"engineError,omitempty"`
}
