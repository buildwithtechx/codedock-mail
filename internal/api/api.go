package api

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/buildwithtechx/codedock-mail/internal/autoconfig"
	"github.com/buildwithtechx/codedock-mail/internal/certs"
	"github.com/buildwithtechx/codedock-mail/internal/dns"
	"github.com/buildwithtechx/codedock-mail/internal/models"
	"github.com/buildwithtechx/codedock-mail/internal/provisioner"
	"github.com/buildwithtechx/codedock-mail/internal/store"
)

type Server struct {
	echo   *echo.Echo
	store  *store.Store
	engine provisioner.Engine
}

func New(s *store.Store, engine provisioner.Engine, token string) *Server {
	e := echo.New()
	if token != "" {
		e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			KeyLookup:  "header:Authorization",
			AuthScheme: "Bearer",
			Validator: func(key string, c echo.Context) (bool, error) {
				return key == token, nil
			},
			Skipper: func(c echo.Context) bool {
				path := c.Request().URL.Path
				return path == "/health" ||
					strings.HasPrefix(path, "/autoconfig/") ||
					strings.HasPrefix(path, "/.well-known/autoconfig/") ||
					strings.HasPrefix(path, "/autodiscover/")
			},
		}))
	}
	srv := &Server{echo: e, store: s, engine: engine}
	e.GET("/health", srv.health)
	e.GET("/api/domains", srv.listDomains)
	e.POST("/api/domains", srv.createDomain)
	e.GET("/api/domains/:id", srv.getDomain)
	e.DELETE("/api/domains/:id", srv.deleteDomain)
	e.GET("/api/domains/:id/mailboxes", srv.listMailboxes)
	e.POST("/api/domains/:id/mailboxes", srv.createMailbox)
	e.DELETE("/api/domains/:domainId/mailboxes/:mailboxId", srv.deleteMailbox)
	e.POST("/api/domains/:id/dkim", srv.createDKIM)
	e.GET("/api/domains/:id/dns", srv.domainDNS)
	e.GET("/api/domains/:id/checks", srv.domainChecks)
	e.GET("/api/certs", srv.listCerts)
	e.POST("/api/certs/probe", srv.probeCert)
	e.GET("/autoconfig/mail/config-v1.1.xml", srv.thunderbirdConfig)
	e.GET("/.well-known/autoconfig/mail/config-v1.1.xml", srv.thunderbirdConfig)
	e.GET("/autodiscover/autodiscover.xml", srv.outlookConfig)
	e.POST("/autodiscover/autodiscover.xml", srv.outlookConfig)
	return srv
}

func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

func (s *Server) health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listDomains(c echo.Context) error {
	domains, err := s.store.ListDomains(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, domains)
}

func (s *Server) createDomain(c echo.Context) error {
	var input models.CreateDomainInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}
	if s.engine != nil {
		if err := s.engine.Provision(input.Name); err != nil {
			return writeError(c, err)
		}
	}
	d, err := s.store.CreateDomain(c.Request().Context(), input.Name)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusCreated, d)
}

func (s *Server) listMailboxes(c echo.Context) error {
	mailboxes, err := s.store.ListMailboxes(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, mailboxes)
}

func (s *Server) createMailbox(c echo.Context) error {
	var input models.CreateMailboxInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	input.LocalPart = strings.TrimSpace(input.LocalPart)
	if input.LocalPart == "" || strings.Contains(input.LocalPart, "@") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "localPart is required and must not contain @"})
	}
	ctx := c.Request().Context()
	d, err := s.store.GetDomain(ctx, c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	credentials := models.MailboxCredentials{}
	if s.engine != nil {
		password, err := provisioner.GeneratePassword()
		if err != nil {
			return writeError(c, err)
		}
		hash, err := provisioner.HashSecret(password)
		if err != nil {
			return writeError(c, err)
		}
		if err := s.engine.AddMailbox(input.LocalPart+"@"+d.Name, hash); err != nil {
			return writeError(c, err)
		}
		credentials.Password = password
	}
	m, err := s.store.CreateMailbox(ctx, d.ID, input.LocalPart)
	if err != nil {
		return writeError(c, err)
	}
	credentials.Mailbox = m
	return c.JSON(http.StatusCreated, credentials)
}

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
	return c.JSON(http.StatusCreated, key)
}

func (s *Server) domainDNS(c echo.Context) error {
	ctx := c.Request().Context()
	d, err := s.store.GetDomain(ctx, c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	response := models.DNSResponse{Domain: d.Name, Records: dns.ForDomain(d.Name, "mail."+d.Name)}
	keys, err := s.store.ListDKIMKeys(ctx, d.ID)
	if err != nil {
		return writeError(c, err)
	}
	for _, key := range keys {
		if key.DNSRecord != "" {
			response.Records = append(response.Records, dns.DKIMRecord(key.Selector, key.DNSRecord))
		}
	}
	if s.engine != nil {
		engineRecords, err := s.engine.DNSRecords(d.Name)
		if err != nil {
			response.EngineError = err.Error()
		} else {
			response.Records = append(response.Records, engineRecords...)
		}
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

func (s *Server) listCerts(c echo.Context) error {
	certs, err := s.store.ListCerts(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, certs)
}

func (s *Server) probeCert(c echo.Context) error {
	var input models.ProbeCertInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	input.Host = strings.TrimSpace(input.Host)
	if input.Host == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "host is required"})
	}
	if input.Port == 0 {
		input.Port = 443
	}
	expiry, err := certs.Probe(c.Request().Context(), input.Host, input.Port)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	cert, err := s.store.UpsertCert(c.Request().Context(), input.Host, expiry)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, cert)
}

func (s *Server) getDomain(c echo.Context) error {
	d, err := s.store.GetDomain(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, d)
}

func (s *Server) deleteDomain(c echo.Context) error {
	ctx := c.Request().Context()
	d, err := s.store.GetDomain(ctx, c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	if s.engine != nil {
		if err := s.engine.RemoveDomain(d.Name); err != nil {
			return writeError(c, err)
		}
	}
	if err := s.store.DeleteDomain(ctx, d.ID); err != nil {
		return writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) deleteMailbox(c echo.Context) error {
	ctx := c.Request().Context()
	m, err := s.store.GetMailbox(ctx, c.Param("mailboxId"))
	if err != nil {
		return writeError(c, err)
	}
	if m.DomainID != c.Param("domainId") {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	if s.engine != nil {
		if err := s.engine.RemoveMailbox(m.Address); err != nil {
			return writeError(c, err)
		}
	}
	if err := s.store.DeleteMailbox(ctx, m.ID); err != nil {
		return writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) thunderbirdConfig(c echo.Context) error {
	email := strings.TrimSpace(c.QueryParam("emailaddress"))
	if email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "emailaddress is required"})
	}
	provider, err := s.lookupProvider(c, email)
	if err != nil {
		return err
	}
	raw, err := autoconfig.ThunderbirdXML(provider)
	if err != nil {
		return writeError(c, err)
	}
	return c.Blob(http.StatusOK, "application/xml", raw)
}

func (s *Server) outlookConfig(c echo.Context) error {
	email := strings.TrimSpace(c.QueryParam("Email"))
	if email == "" && c.Request().Method == http.MethodPost {
		raw, err := io.ReadAll(io.LimitReader(c.Request().Body, 1<<20))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}
		email = autoconfig.OutlookRequestEmail(raw)
	}
	if email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "email address is required"})
	}
	provider, err := s.lookupProvider(c, email)
	if err != nil {
		return err
	}
	raw, err := autoconfig.OutlookXML(provider, email)
	if err != nil {
		return writeError(c, err)
	}
	return c.Blob(http.StatusOK, "application/xml", raw)
}

func (s *Server) lookupProvider(c echo.Context, email string) (autoconfig.Provider, error) {
	domain, err := autoconfig.DomainOf(email)
	if err != nil {
		return autoconfig.Provider{}, c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid email address"})
	}
	d, err := s.store.GetDomainByName(c.Request().Context(), domain)
	if err != nil {
		return autoconfig.Provider{}, writeError(c, err)
	}
	return autoconfig.Provider{Domain: d.Name, DisplayName: d.Name + " Mail", MailHost: "mail." + d.Name}, nil
}

func writeError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, store.ErrExists), errors.Is(err, provisioner.ErrEngineExists):
		return c.JSON(http.StatusConflict, map[string]string{"error": "already exists"})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}
