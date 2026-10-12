package dns

import (
	"context"
	"net"
	"strings"
	"time"
)

type Check struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

type Report struct {
	Domain string  `json:"domain"`
	Checks []Check `json:"checks"`
}

func VerifyDomain(ctx context.Context, domain, dkimSelector string) Report {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	report := Report{Domain: domain, Checks: []Check{}}
	mxs, err := net.DefaultResolver.LookupMX(ctx, domain)
	if err != nil || len(mxs) == 0 {
		report.Checks = append(report.Checks, Check{Name: "mx", Passed: false, Detail: "no MX records"})
	} else {
		hosts := make([]string, 0, len(mxs))
		for _, mx := range mxs {
			hosts = append(hosts, mx.Host)
		}
		report.Checks = append(report.Checks, Check{Name: "mx", Passed: true, Detail: strings.Join(hosts, ", ")})
	}
	report.Checks = append(report.Checks, checkTXT(ctx, domain, "spf", "v=spf1"))
	report.Checks = append(report.Checks, checkTXT(ctx, "_dmarc."+domain, "dmarc", "v=DMARC1"))
	if dkimSelector != "" {
		report.Checks = append(report.Checks, checkTXT(ctx, dkimSelector+"._domainkey."+domain, "dkim", "v=DKIM1"))
	}
	return report
}

func checkTXT(ctx context.Context, name, checkName, prefix string) Check {
	txts, err := net.DefaultResolver.LookupTXT(ctx, name)
	if err != nil {
		return Check{Name: checkName, Passed: false, Detail: "lookup failed"}
	}
	for _, txt := range txts {
		if strings.HasPrefix(txt, prefix) {
			return Check{Name: checkName, Passed: true, Detail: txt}
		}
	}
	return Check{Name: checkName, Passed: false, Detail: "record missing"}
}
