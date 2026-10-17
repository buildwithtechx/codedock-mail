package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/buildwithtechx/codedock-mail/internal/imap"
)

type MailboxSSEEvent struct {
	Folder    string `json:"folder"`
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
}

func (s *Server) mailIdle(c echo.Context) error {
	folder := c.QueryParam("view")
	if folder == "" {
		folder = c.QueryParam("folder")
	}
	folder = imap.NormalizeFolderSlug(folder)

	res := c.Response()
	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")
	// OpenResty/Nginx must forward each notification as it arrives without buffering.
	res.Header().Set("X-Accel-Buffering", "no")
	res.WriteHeader(http.StatusOK)
	res.Flush()

	connectedPayload, _ := json.Marshal(MailboxSSEEvent{
		Folder:    folder,
		Type:      "connected",
		Timestamp: time.Now().Unix(),
	})
	if _, err := fmt.Fprintf(res, "event: mailbox\ndata: %s\n\n", connectedPayload); err != nil {
		return err
	}
	res.Flush()

	ctx := c.Request().Context()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-ticker.C:
			heartbeat, _ := json.Marshal(MailboxSSEEvent{
				Folder:    folder,
				Type:      "heartbeat",
				Timestamp: t.Unix(),
			})
			if _, err := fmt.Fprintf(res, "event: mailbox\ndata: %s\n\n", heartbeat); err != nil {
				return nil
			}
			res.Flush()
		}
	}
}
