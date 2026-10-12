package dns

type Record struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func ForDomain(domain string) []Record {
	return []Record{
		{Type: "MX", Name: "@", Value: "10 mail." + domain},
		{Type: "TXT", Name: "@", Value: "v=spf1 mx -all"},
	}
}
