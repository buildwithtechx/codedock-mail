package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (s *Server) listAudit(c echo.Context) error {
	limit := 50
	if raw := strings.TrimSpace(c.QueryParam("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	entries, err := s.store.ListAudit(c.Request().Context(), limit)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, entries)
}
