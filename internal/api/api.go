package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/buildwithtechx/codedock-mail/internal/models"
	"github.com/buildwithtechx/codedock-mail/internal/store"
)

type Server struct {
	echo  *echo.Echo
	store *store.Store
}

func New(s *store.Store) *Server {
	e := echo.New()
	srv := &Server{echo: e, store: s}
	e.GET("/health", srv.health)
	e.GET("/api/domains", srv.listDomains)
	e.POST("/api/domains", srv.createDomain)
	e.GET("/api/domains/:id/mailboxes", srv.listMailboxes)
	e.POST("/api/domains/:id/mailboxes", srv.createMailbox)
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
	m, err := s.store.CreateMailbox(c.Request().Context(), c.Param("id"), input.LocalPart)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusCreated, m)
}

func writeError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, store.ErrExists):
		return c.JSON(http.StatusConflict, map[string]string{"error": "already exists"})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}
