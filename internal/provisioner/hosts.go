package provisioner

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var hostnameRegex = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,252}$`)

type ExternalWebmailBackend struct {
	IMAPHost string `json:"imapHost"`
	IMAPPort int    `json:"imapPort"`
	SMTPHost string `json:"smtpHost"`
	SMTPPort int    `json:"smtpPort"`
	UseTLS   bool   `json:"useTls"`
	IsLocal  bool   `json:"isLocal"`
}

func IsValidAddressOrGateway(addr string) bool {
	if addr == "host-gateway" {
		return true
	}
	if ip := net.ParseIP(addr); ip != nil {
		return true
	}
	// Try bracketed IPv6
	if strings.HasPrefix(addr, "[") && strings.HasSuffix(addr, "]") {
		if ip := net.ParseIP(addr[1 : len(addr)-1]); ip != nil {
			return true
		}
	}
	// Verify URL format with IPv6 host
	if strings.Contains(addr, ":") {
		u, err := url.Parse(fmt.Sprintf("http://[%s]/", strings.Trim(addr, "[]")))
		if err == nil && u != nil && strings.HasPrefix(u.Hostname(), "[") {
			return true
		}
	}
	return false
}

func ParseContainerHosts(raw any, networkMode string) (map[string]string, error) {
	if raw == nil {
		return nil, nil
	}

	if strings.HasPrefix(strings.ToLower(networkMode), "container:") {
		return nil, fmt.Errorf("extra_hosts cannot be used with a shared container network namespace")
	}

	entries := make([][2]string, 0)

	switch val := raw.(type) {
	case []string:
		for _, item := range val {
			sep := strings.IndexAny(item, "=:")
			if sep < 1 {
				return nil, fmt.Errorf("extra_hosts entries must be hostname=address or hostname:address")
			}
			entries = append(entries, [2]string{item[:sep], item[sep+1:]})
		}
	case map[string]string:
		for k, v := range val {
			entries = append(entries, [2]string{k, v})
		}
	case map[string]any:
		for k, v := range val {
			strVal, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("extra_hosts requires string addresses")
			}
			entries = append(entries, [2]string{k, strVal})
		}
	default:
		return nil, fmt.Errorf("extra_hosts must be a hostname/address map or list")
	}

	if len(entries) > 100 {
		return nil, fmt.Errorf("extra_hosts supports at most 100 aliases")
	}

	result := make(map[string]string)
	for _, pair := range entries {
		host := strings.ToLower(strings.TrimSpace(pair[0]))
		addr := strings.TrimSpace(pair[1])
		addr = strings.TrimPrefix(addr, "[")
		addr = strings.TrimSuffix(addr, "]")

		if !hostnameRegex.MatchString(host) {
			return nil, fmt.Errorf("invalid hostname in extra_hosts: %s", host)
		}
		if len(addr) > 64 || !IsValidAddressOrGateway(addr) {
			return nil, fmt.Errorf("extra_hosts addresses must be IPv4, IPv6 or host-gateway: %s", addr)
		}
		if existing, exists := result[host]; exists && existing != addr {
			return nil, fmt.Errorf("extra_hosts declares conflicting addresses for hostname %s", host)
		}
		result[host] = addr
	}

	return result, nil
}

func ConfigureExternalWebmailInstall(backend ExternalWebmailBackend) (map[string]string, error) {
	if backend.IMAPHost == "" || backend.SMTPHost == "" {
		return nil, fmt.Errorf("both IMAP and SMTP hosts are required")
	}

	extraHosts := make(map[string]string)
	if backend.IsLocal {
		extraHosts[backend.IMAPHost] = "host-gateway"
		if backend.SMTPHost != backend.IMAPHost {
			extraHosts[backend.SMTPHost] = "host-gateway"
		}
	}

	return ParseContainerHosts(extraHosts, "")
}
