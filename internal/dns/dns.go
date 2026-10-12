package dns

type Record struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

func ForDomain(domain, mailHost string) []Record {
	return []Record{
		{Type: "MX", Name: "@", Value: "10 " + mailHost, Source: "local"},
		{Type: "TXT", Name: "@", Value: "v=spf1 mx -all", Source: "local"},
		{Type: "TXT", Name: "_dmarc", Value: "v=DMARC1; p=quarantine; rua=mailto:postmaster@" + domain, Source: "local"},
	}
}

func DKIMRecord(selector, txt string) Record {
	return Record{Type: "TXT", Name: selector + "._domainkey", Value: txt, Source: "local"}
}
