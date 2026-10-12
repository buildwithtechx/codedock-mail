package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/buildwithtechx/codedock-mail/apps/admin"
	"github.com/buildwithtechx/codedock-mail/internal/backups"
	"github.com/buildwithtechx/codedock-mail/internal/models"
	"github.com/buildwithtechx/codedock-mail/internal/provisioner"
	"github.com/buildwithtechx/codedock-mail/internal/store"
)

type Server struct {
	echo        *echo.Echo
	store       *store.Store
	engine      provisioner.Engine
	dsn         string
	backupDir   string
	mailDataDir string
}

func (s *Server) SetDatabase(dsn, backupDir string) {
	s.dsn = dsn
	s.backupDir = backupDir
}

func (s *Server) SetMailDataDir(dir string) {
	s.mailDataDir = dir
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
				return !strings.HasPrefix(c.Request().URL.Path, "/api/")
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
	e.PUT("/api/domains/:domainId/mailboxes/:mailboxId/quota", srv.updateQuota)
	e.POST("/api/domains/:id/dkim", srv.createDKIM)
	e.GET("/api/domains/:id/dns", srv.domainDNS)
	e.GET("/api/domains/:id/checks", srv.domainChecks)
	e.GET("/api/certs", srv.listCerts)
	e.POST("/api/certs/probe", srv.probeCert)
	e.GET("/api/domains/:domainId/mailboxes/:mailboxId/sieve", srv.getSieve)
	e.PUT("/api/domains/:domainId/mailboxes/:mailboxId/sieve", srv.putSieve)
	e.GET("/api/backups", srv.listBackups)
	e.POST("/api/backups", srv.createBackup)
	e.POST("/api/backups/:id/restore", srv.restoreBackup)
	e.DELETE("/api/backups/:id", srv.deleteBackup)
	e.GET("/api/audit", srv.listAudit)
	e.GET("/autoconfig/mail/config-v1.1.xml", srv.thunderbirdConfig)
	e.GET("/.well-known/autoconfig/mail/config-v1.1.xml", srv.thunderbirdConfig)
	e.GET("/autodiscover/autodiscover.xml", srv.outlookConfig)
	e.POST("/autodiscover/autodiscover.xml", srv.outlookConfig)
	return srv
}

func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

func (s *Server) RegisterAdmin() {
	admin.RegisterHandlers(s.echo)
}

func (s *Server) health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) mailboxInDomain(ctx context.Context, domainID, mailboxID string) (models.Mailbox, error) {
	m, err := s.store.GetMailbox(ctx, mailboxID)
	if err != nil {
		return models.Mailbox{}, err
	}
	if m.DomainID != domainID {
		return models.Mailbox{}, store.ErrNotFound
	}
	return m, nil
}

func (s *Server) backupDirOrDefault() string {
	if s.backupDir != "" {
		return s.backupDir
	}
	return backups.Dir()
}

func (s *Server) audit(ctx context.Context, action, resourceType, resourceID, detail string) {
	_, _ = s.store.Log(ctx, models.AuditEntry{
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Detail:       detail,
	})
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
