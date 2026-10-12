package api

import (
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/buildwithtechx/codedock-mail/internal/autoconfig"
)

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
