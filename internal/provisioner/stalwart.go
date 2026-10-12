package provisioner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/buildwithtechx/codedock-mail/internal/dns"
)

var ErrEngineExists = errors.New("already exists")

type StalwartEngine struct {
	baseURL  string
	username string
	secret   string
	client   *http.Client
}

func NewStalwart(baseURL, username, secret string) *StalwartEngine {
	return &StalwartEngine{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		username: username,
		secret:   secret,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

type stalwartEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

func (e *StalwartEngine) Provision(domain string) error {
	_, err := e.do(http.MethodPost, "/api/principal", map[string]string{
		"type": "domain",
		"name": domain,
	})
	return err
}

func (e *StalwartEngine) AddMailbox(address, secretHash string, quotaBytes int64) error {
	_, err := e.do(http.MethodPost, "/api/principal", map[string]any{
		"type":    "individual",
		"name":    address,
		"secrets": []string{secretHash},
		"emails":  []string{address},
		"quota":   quotaBytes,
	})
	return err
}

func (e *StalwartEngine) UpdateMailboxQuota(address string, quotaBytes int64) error {
	_, err := e.do(http.MethodPatch, "/api/principal/"+url.PathEscape(address), map[string]any{
		"quota": quotaBytes,
	})
	return err
}

func (e *StalwartEngine) RemoveMailbox(address string) error {
	_, err := e.do(http.MethodDelete, "/api/principal/"+url.PathEscape(address), nil)
	return err
}

func (e *StalwartEngine) RemoveDomain(domain string) error {
	_, err := e.do(http.MethodDelete, "/api/principal/"+url.PathEscape(domain), nil)
	return err
}

func (e *StalwartEngine) CreateDKIM(domain, selector, algorithm string) (DKIMResult, error) {
	raw, err := e.do(http.MethodPost, "/api/dkim", map[string]string{
		"algorithm": algorithm,
		"domain":    domain,
		"selector":  selector,
	})
	if err != nil {
		return DKIMResult{}, err
	}
	result := DKIMResult{Selector: selector}
	if len(raw) > 0 {
		var data struct {
			Selector  string `json:"selector"`
			PublicKey string `json:"public_key"`
			DNSRecord string `json:"dns_record"`
		}
		if err := json.Unmarshal(raw, &data); err != nil {
			return DKIMResult{}, fmt.Errorf("stalwart: decode dkim response: %w", err)
		}
		if data.Selector != "" {
			result.Selector = data.Selector
		}
		result.PublicKey = data.PublicKey
		result.DNSRecord = data.DNSRecord
	}
	if result.DNSRecord == "" {
		records, err := e.DNSRecords(domain)
		if err != nil {
			return DKIMResult{}, err
		}
		for _, record := range records {
			if record.Type == "TXT" && strings.HasPrefix(record.Name, result.Selector+"._domainkey.") {
				result.DNSRecord = record.Value
				result.PublicKey = dkimPublicKey(record.Value)
				break
			}
		}
		if result.DNSRecord == "" {
			return DKIMResult{}, fmt.Errorf("stalwart: dkim key created but no dns record found")
		}
	}
	return result, nil
}

func dkimPublicKey(txt string) string {
	for _, part := range strings.Split(txt, ";") {
		part = strings.TrimSpace(part)
		if value, ok := strings.CutPrefix(part, "p="); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (e *StalwartEngine) DNSRecords(domain string) ([]dns.Record, error) {
	raw, err := e.do(http.MethodGet, "/api/dns/records/"+url.PathEscape(domain), nil)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, fmt.Errorf("stalwart: decode dns records: %w", err)
		}
	}
	records := []dns.Record{}
	for _, item := range items {
		typ := recordField(item, "type")
		name := recordField(item, "name", "host")
		value := recordField(item, "content", "value", "target", "data")
		if typ == "" || name == "" || value == "" {
			continue
		}
		records = append(records, dns.Record{Type: strings.ToUpper(typ), Name: name, Value: value, Source: "engine"})
	}
	return records, nil
}

func (e *StalwartEngine) Health() error {
	_, err := e.do(http.MethodGet, "/api/principal?limit=1", nil)
	return err
}

func (e *StalwartEngine) SieveAddr() string {
	return net.JoinHostPort(e.engineHost(), "4190")
}

func (e *StalwartEngine) ImapAddr() string {
	return net.JoinHostPort(e.engineHost(), "993")
}

func (e *StalwartEngine) engineHost() string {
	if parsed, err := url.Parse(e.baseURL); err == nil && parsed.Hostname() != "" {
		return parsed.Hostname()
	}
	return e.baseURL
}

func (e *StalwartEngine) ReloadCertificates() error {
	_, err := e.do(http.MethodGet, "/api/reload/certificate", nil)
	return err
}

func (e *StalwartEngine) do(method, path string, body any) (json.RawMessage, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("stalwart: encode request: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, e.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("stalwart: build request: %w", err)
	}
	req.SetBasicAuth(e.username, e.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stalwart: request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("stalwart: read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("stalwart: %s %s: status %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var env stalwartEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("stalwart: decode response: %w", err)
	}
	if env.Error != "" {
		if isExistsError(env.Error) {
			return nil, ErrEngineExists
		}
		return nil, fmt.Errorf("stalwart: %s", env.Error)
	}
	return env.Data, nil
}

func isExistsError(message string) bool {
	lowered := strings.ToLower(message)
	return strings.Contains(lowered, "exist") ||
		strings.Contains(lowered, "duplicate") ||
		strings.Contains(lowered, "conflict") ||
		strings.Contains(lowered, "already")
}

func recordField(item map[string]any, keys ...string) string {
	for _, key := range keys {
		for actual, value := range item {
			if strings.EqualFold(actual, key) {
				if text, ok := value.(string); ok && text != "" {
					return text
				}
			}
		}
	}
	return ""
}
