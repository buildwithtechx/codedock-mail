package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/buildwithtechx/codedock-mail/internal/dns"
	"github.com/buildwithtechx/codedock-mail/internal/models"
	"github.com/buildwithtechx/codedock-mail/internal/provisioner"
	"github.com/buildwithtechx/codedock-mail/internal/store"
)

func (s *Server) createDKIM(c echo.Context) error {
	if s.engine == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "engine not configured"})
	}
	var input models.CreateDKIMInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	input.Selector = strings.TrimSpace(input.Selector)
	if input.Selector == "" {
		input.Selector = "mail"
	}
	if strings.ContainsAny(input.Selector, " @.") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "selector must be a single dns label"})
	}
	input.Algorithm = strings.TrimSpace(input.Algorithm)
	if input.Algorithm == "" {
		input.Algorithm = "Ed25519"
	}
	ctx := c.Request().Context()
	d, err := s.store.GetDomain(ctx, c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	result, err := s.engine.CreateDKIM(d.Name, input.Selector, input.Algorithm)
	if err != nil {
		return writeError(c, err)
	}
	key, err := s.store.CreateDKIMKey(ctx, models.DKIMKey{
		DomainID:  d.ID,
		Selector:  result.Selector,
		Algorithm: input.Algorithm,
		PublicKey: result.PublicKey,
		DNSRecord: result.DNSRecord,
	})
	if err != nil {
		return writeError(c, err)
	}
	s.audit(ctx, "dkim.create", "domain", d.ID, result.Selector)
	return c.JSON(http.StatusCreated, key)
}

func (s *Server) domainDNS(c echo.Context) error {
	ctx := c.Request().Context()
	d, err := s.store.GetDomain(ctx, c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	response := models.DNSResponse{Domain: d.Name}
	local := dns.ForDomain(d.Name, "mail."+d.Name)
	keys, err := s.store.ListDKIMKeys(ctx, d.ID)
	if err != nil {
		return writeError(c, err)
	}
	for _, key := range keys {
		if key.DNSRecord != "" {
			local = append(local, dns.DKIMRecord(key.Selector, key.DNSRecord))
		}
	}
	if s.engine != nil {
		engineRecords, err := s.engine.DNSRecords(d.Name)
		if err != nil {
			response.EngineError = err.Error()
			response.Records = local
		} else {
			response.Records = dns.Merge(d.Name, local, engineRecords)
		}
	} else {
		response.Records = local
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) domainChecks(c echo.Context) error {
	ctx := c.Request().Context()
	d, err := s.store.GetDomain(ctx, c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	selector := ""
	if key, err := s.store.GetDKIMKey(ctx, d.ID); err == nil {
		selector = key.Selector
	} else if !errors.Is(err, store.ErrNotFound) {
		return writeError(c, err)
	}
	report := dns.VerifyDomain(ctx, d.Name, selector)
	if err := provisioner.CheckPort25Egress(ctx); err != nil {
		report.Checks = append(report.Checks, dns.Check{Name: "port25", Passed: false, Detail: err.Error()})
	} else {
		report.Checks = append(report.Checks, dns.Check{Name: "port25", Passed: true, Detail: "tcp connect to port 25 succeeded"})
	}
	if ip := strings.TrimSpace(c.QueryParam("ip")); ip != "" {
		if ptr, err := provisioner.CheckPTR(ctx, ip); err != nil {
			report.Checks = append(report.Checks, dns.Check{Name: "ptr", Passed: false, Detail: err.Error()})
		} else {
			report.Checks = append(report.Checks, dns.Check{Name: "ptr", Passed: true, Detail: ptr})
		}
	}
	return c.JSON(http.StatusOK, report)
}
