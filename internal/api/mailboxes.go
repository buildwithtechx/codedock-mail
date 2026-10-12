package api

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/buildwithtechx/codedock-mail/internal/imap"
	"github.com/buildwithtechx/codedock-mail/internal/models"
	"github.com/buildwithtechx/codedock-mail/internal/provisioner"
)

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
	if input.QuotaMB < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "quotaMb must not be negative"})
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
		if err := s.engine.AddMailbox(input.LocalPart+"@"+d.Name, hash, input.QuotaMB*1024*1024); err != nil {
			return writeError(c, err)
		}
		credentials.Password = password
	}
	m, err := s.store.CreateMailbox(ctx, d.ID, input.LocalPart, input.QuotaMB)
	if err != nil {
		return writeError(c, err)
	}
	credentials.Mailbox = m
	s.audit(ctx, "mailbox.create", "mailbox", m.ID, m.Address)
	return c.JSON(http.StatusCreated, credentials)
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
	s.audit(ctx, "mailbox.delete", "mailbox", m.ID, m.Address)
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) updateQuota(c echo.Context) error {
	var input models.UpdateQuotaInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	if input.QuotaMB < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "quotaMb must not be negative"})
	}
	ctx := c.Request().Context()
	m, err := s.mailboxInDomain(ctx, c.Param("domainId"), c.Param("mailboxId"))
	if err != nil {
		return writeError(c, err)
	}
	if s.engine != nil {
		if err := s.engine.UpdateMailboxQuota(m.Address, input.QuotaMB*1024*1024); err != nil {
			return writeError(c, err)
		}
	}
	updated, err := s.store.UpdateMailboxQuota(ctx, m.ID, input.QuotaMB)
	if err != nil {
		return writeError(c, err)
	}
	s.audit(ctx, "mailbox.quota", "mailbox", m.ID, m.Address)
	return c.JSON(http.StatusOK, updated)
}

func (s *Server) getUsage(c echo.Context) error {
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
	used, err := imap.GetQuota(ctx, s.engine.ImapAddr(), m.Address, password)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, models.MailboxUsage{Address: m.Address, UsedBytes: used, QuotaMB: m.QuotaMB})
}
