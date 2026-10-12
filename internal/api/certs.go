package api

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/buildwithtechx/codedock-mail/internal/certs"
	"github.com/buildwithtechx/codedock-mail/internal/models"
)

func (s *Server) listCerts(c echo.Context) error {
	tracked, err := s.store.ListCerts(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, tracked)
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
