package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/buildwithtechx/codedock-mail/internal/backups"
	"github.com/buildwithtechx/codedock-mail/internal/provisioner"
)

func (s *Server) listBackups(c echo.Context) error {
	manifests, err := backups.List(s.backupDirOrDefault())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, manifests)
}

func (s *Server) createBackup(c echo.Context) error {
	if s.dsn == "" {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
	}
	manifest, err := backups.Create(c.Request().Context(), s.dsn, s.backupDirOrDefault(), s.mailDataDir)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	s.audit(c.Request().Context(), "backup.create", "backup", manifest.ID, manifest.Path)
	return c.JSON(http.StatusCreated, manifest)
}

func (s *Server) restoreBackup(c echo.Context) error {
	if s.dsn == "" {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
	}
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing backup id"})
	}
	if err := backups.Restore(c.Request().Context(), s.dsn, s.backupDirOrDefault(), id); err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	s.audit(c.Request().Context(), "backup.restore", "backup", id, "")
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) restoreMail(c echo.Context) error {
	if s.engine == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "engine not configured"})
	}
	if s.mailDataDir == "" {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "mail data dir not configured"})
	}
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing backup id"})
	}
	manifests, err := backups.List(s.backupDirOrDefault())
	if err != nil {
		return writeError(c, err)
	}
	dataPath := ""
	found := false
	for _, manifest := range manifests {
		if manifest.ID == id {
			dataPath = manifest.DataPath
			found = true
			break
		}
	}
	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	if dataPath == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "backup has no mail archive"})
	}
	ctx := c.Request().Context()
	stopped := func() bool { return provisioner.TCPPortClosed(ctx, s.engine.ImapAddr()) }
	restorePoint, err := backups.RestoreMailData(s.mailDataDir, dataPath, stopped)
	if err != nil {
		if errors.Is(err, backups.ErrEngineRunning) {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	s.audit(ctx, "backup.restore-mail", "backup", id, restorePoint)
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "restorePoint": restorePoint})
}

func (s *Server) deleteBackup(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing backup id"})
	}
	if err := backups.Delete(s.backupDirOrDefault(), id); err != nil {
		return writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
