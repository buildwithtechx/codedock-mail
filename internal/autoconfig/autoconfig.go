package autoconfig

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type Provider struct {
	Domain      string
	DisplayName string
	MailHost    string
}

type tbServer struct {
	Type       string `xml:"type,attr"`
	Hostname   string `xml:"hostname"`
	Port       int    `xml:"port"`
	SocketType string `xml:"socketType"`
	Auth       string `xml:"authentication"`
	Username   string `xml:"username"`
}

type tbProvider struct {
	ID       string   `xml:"id,attr"`
	Domain   string   `xml:"domain"`
	Display  string   `xml:"displayName"`
	Incoming tbServer `xml:"incomingServer"`
	Outgoing tbServer `xml:"outgoingServer"`
}

type tbConfig struct {
	XMLName  xml.Name   `xml:"clientConfig"`
	Version  string     `xml:"version,attr"`
	Provider tbProvider `xml:"emailProvider"`
}

func ThunderbirdXML(provider Provider) ([]byte, error) {
	config := tbConfig{
		Version: "1.1",
		Provider: tbProvider{
			ID:      provider.Domain,
			Domain:  provider.Domain,
			Display: provider.DisplayName,
			Incoming: tbServer{
				Type: "imap", Hostname: provider.MailHost, Port: 993,
				SocketType: "SSL", Auth: "password-cleartext", Username: "%EMAILADDRESS%",
			},
			Outgoing: tbServer{
				Type: "smtp", Hostname: provider.MailHost, Port: 587,
				SocketType: "STARTTLS", Auth: "password-cleartext", Username: "%EMAILADDRESS%",
			},
		},
	}
	raw, err := xml.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal thunderbird config: %w", err)
	}
	return append([]byte(xml.Header), raw...), nil
}

type oxProtocol struct {
	Type      string `xml:"Type"`
	Server    string `xml:"Server"`
	Port      int    `xml:"Port"`
	DomainReq string `xml:"DomainRequired"`
	LoginName string `xml:"LoginName"`
	SSL       string `xml:"SSL"`
	AuthReq   string `xml:"AuthRequired"`
}

type oxAccount struct {
	AccountType string       `xml:"AccountType"`
	Action      string       `xml:"Action"`
	Protocols   []oxProtocol `xml:"Protocol"`
}

type oxResponse struct {
	Account oxAccount `xml:"Account"`
}

type oxDiscover struct {
	XMLName  xml.Name   `xml:"Autodiscover"`
	Xmlns    string     `xml:"xmlns,attr"`
	Response oxResponse `xml:"Response"`
}

func OutlookXML(provider Provider, email string) ([]byte, error) {
	discover := oxDiscover{
		Xmlns: "http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a",
		Response: oxResponse{
			Account: oxAccount{
				AccountType: "email",
				Action:      "settings",
				Protocols: []oxProtocol{
					{Type: "IMAP", Server: provider.MailHost, Port: 993, DomainReq: "off", LoginName: email, SSL: "on", AuthReq: "on"},
					{Type: "SMTP", Server: provider.MailHost, Port: 587, DomainReq: "off", LoginName: email, SSL: "on", AuthReq: "on"},
				},
			},
		},
	}
	raw, err := xml.MarshalIndent(discover, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal outlook config: %w", err)
	}
	return append([]byte(xml.Header), raw...), nil
}

type oxRequest struct {
	Request struct {
		EMailAddress string `xml:"EMailAddress"`
	} `xml:"Request"`
}

func OutlookRequestEmail(raw []byte) string {
	var request oxRequest
	if err := xml.Unmarshal(raw, &request); err != nil {
		return ""
	}
	return strings.TrimSpace(request.Request.EMailAddress)
}

func DomainOf(email string) (string, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return "", fmt.Errorf("invalid email address")
	}
	return parts[1], nil
}
