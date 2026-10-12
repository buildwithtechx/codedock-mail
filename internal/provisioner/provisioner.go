package provisioner

type Engine interface {
	Provision(domain string) error
	AddMailbox(address string) error
	Health() error
}

type Provisioner struct {
	engine Engine
}

func New(engine Engine) *Provisioner {
	return &Provisioner{engine: engine}
}
