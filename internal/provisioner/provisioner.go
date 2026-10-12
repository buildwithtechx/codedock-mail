package provisioner

import "github.com/buildwithtechx/codedock-mail/internal/dns"

type DKIMResult struct {
	Selector  string `json:"selector"`
	PublicKey string `json:"publicKey"`
	DNSRecord string `json:"dnsRecord"`
}

type Engine interface {
	Provision(domain string) error
	RemoveDomain(domain string) error
	AddMailbox(address, secretHash string) error
	RemoveMailbox(address string) error
	CreateDKIM(domain, selector, algorithm string) (DKIMResult, error)
	DNSRecords(domain string) ([]dns.Record, error)
	ReloadCertificates() error
	Health() error
}

type Provisioner struct {
	engine Engine
}

func New(engine Engine) *Provisioner {
	return &Provisioner{engine: engine}
}
