package api

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/buildwithtechx/codedock-mail/internal/models"
	"github.com/buildwithtechx/codedock-mail/internal/sieve"
)

func (s *Server) getSieve(c echo.Context) error {
	if s.engine == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "engine not configured"})
	}
	password := strings.TrimSpace(c.Request().Header.Get("X-Mailbox-Password"))
	if password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "X-Mailbox-Password header is required"})
	}
	ctx := c.Request().Context()
	m, err := s.mailboxInDomain(ctx, c.Param("domainId"), c.Param("mailboxId"))
	if err != nil {
		return writeError(c, err)
	}
	scripts, content, err := sieve.GetActive(s.engine.SieveAddr(), m.Address, password)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"address": m.Address, "scripts": scripts, "activeScript": content})
}

func (s *Server) putSieve(c echo.Context) error {
	if s.engine == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "engine not configured"})
	}
	var input models.SieveInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	input.Password = strings.TrimSpace(input.Password)
	input.Name = strings.TrimSpace(input.Name)
	if input.Password == "" || input.Name == "" || input.Script == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "password, name, and script are required"})
	}
	ctx := c.Request().Context()
	m, err := s.mailboxInDomain(ctx, c.Param("domainId"), c.Param("mailboxId"))
	if err != nil {
		return writeError(c, err)
	}
	if err := sieve.SetActiveScript(s.engine.SieveAddr(), m.Address, input.Password, input.Name, input.Script); err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	s.audit(ctx, "sieve.update", "mailbox", m.ID, m.Address)
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
