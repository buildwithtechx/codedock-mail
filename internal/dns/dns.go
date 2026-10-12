package dns

import "strings"

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

func Merge(domain string, local, engine []Record) []Record {
	seen := map[string]bool{}
	merged := make([]Record, 0, len(local)+len(engine))
	for _, record := range engine {
		seen[recordKey(domain, record)] = true
		merged = append(merged, record)
	}
	for _, record := range local {
		if seen[recordKey(domain, record)] {
			continue
		}
		merged = append(merged, record)
	}
	return merged
}

func recordKey(domain string, record Record) string {
	base := strings.ToLower(strings.TrimSuffix(domain, "."))
	name := strings.ToLower(strings.TrimSpace(record.Name))
	if name == "@" {
		name = base
	}
	name = strings.TrimSuffix(name, ".")
	if name != base && !strings.HasSuffix(name, "."+base) {
		name = name + "." + base
	}
	return strings.ToUpper(record.Type) + " " + name
}
