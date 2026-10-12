package certs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

const DefaultInterval = time.Hour
const DefaultWarnBefore = 30 * 24 * time.Hour

type CertStore interface {
	ListExpiring(ctx context.Context, before time.Time) ([]models.Cert, error)
	UpsertCert(ctx context.Context, hostname string, expiresAt time.Time) (models.Cert, error)
}

type Reloader interface {
	ReloadCertificates() error
}

type Report struct {
	CheckedAt   time.Time     `json:"checkedAt"`
	Expiring    []models.Cert `json:"expiring"`
	Reloaded    bool          `json:"reloaded"`
	ReloadError string        `json:"reloadError,omitempty"`
}

type Watcher struct {
	store      CertStore
	engine     Reloader
	interval   time.Duration
	warnBefore time.Duration
}

func NewWatcher(store CertStore, engine Reloader, interval, warnBefore time.Duration) *Watcher {
	return &Watcher{store: store, engine: engine, interval: interval, warnBefore: warnBefore}
}

func (w *Watcher) Run(ctx context.Context) error {
	if _, err := w.CheckOnce(ctx); err != nil {
		slog.Error("cert check failed", "error", err)
	}
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			report, err := w.CheckOnce(ctx)
			if err != nil {
				slog.Error("cert check failed", "error", err)
				continue
			}
			if len(report.Expiring) > 0 {
				slog.Warn("certificates near expiry", "count", len(report.Expiring), "reloaded", report.Reloaded)
			}
		}
	}
}

func (w *Watcher) CheckOnce(ctx context.Context) (Report, error) {
	report := Report{CheckedAt: time.Now().UTC(), Expiring: []models.Cert{}}
	expiring, err := w.store.ListExpiring(ctx, time.Now().Add(w.warnBefore))
	if err != nil {
		return report, fmt.Errorf("list expiring certs: %w", err)
	}
	report.Expiring = expiring
	if len(expiring) == 0 || w.engine == nil {
		return report, nil
	}
	if err := w.engine.ReloadCertificates(); err != nil {
		report.ReloadError = err.Error()
		return report, nil
	}
	report.Reloaded = true
	return report, nil
}
