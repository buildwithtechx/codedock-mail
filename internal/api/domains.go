package api

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

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
	s.audit(c.Request().Context(), "domain.create", "domain", d.ID, d.Name)
	return c.JSON(http.StatusCreated, d)
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
	s.audit(ctx, "domain.delete", "domain", d.ID, d.Name)
	return c.NoContent(http.StatusNoContent)
}
