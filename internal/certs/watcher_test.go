package certs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

type fakeStore struct {
	certs    []models.Cert
	upserted []models.Cert
}

func (f *fakeStore) ListExpiring(ctx context.Context, before time.Time) ([]models.Cert, error) {
	expiring := []models.Cert{}
	for _, cert := range f.certs {
		if cert.ExpiresAt.Before(before) {
			expiring = append(expiring, cert)
		}
	}
	return expiring, nil
}

func (f *fakeStore) UpsertCert(ctx context.Context, hostname string, expiresAt time.Time) (models.Cert, error) {
	cert := models.Cert{Hostname: hostname, ExpiresAt: expiresAt}
	f.upserted = append(f.upserted, cert)
	return cert, nil
}

type fakeEngine struct {
	reloads int
	err     error
}

func (f *fakeEngine) ReloadCertificates() error {
	f.reloads++
	return f.err
}

func TestCheckOnceReloadsWhenExpiring(t *testing.T) {
	store := &fakeStore{certs: []models.Cert{{Hostname: "mail.example.com", ExpiresAt: time.Now().Add(time.Hour)}}}
	engine := &fakeEngine{}
	watcher := NewWatcher(store, engine, time.Hour, 30*24*time.Hour)
	report, err := watcher.CheckOnce(context.Background())
	if err != nil {
		t.Fatalf("check once: %v", err)
	}
	if len(report.Expiring) != 1 || !report.Reloaded || engine.reloads != 1 {
		t.Fatalf("expected one expiring cert and one reload, got %+v", report)
	}
}

func TestCheckOnceSkipsReloadWhenHealthy(t *testing.T) {
	store := &fakeStore{certs: []models.Cert{{Hostname: "mail.example.com", ExpiresAt: time.Now().Add(90 * 24 * time.Hour)}}}
	engine := &fakeEngine{}
	watcher := NewWatcher(store, engine, time.Hour, 30*24*time.Hour)
	report, err := watcher.CheckOnce(context.Background())
	if err != nil {
		t.Fatalf("check once: %v", err)
	}
	if len(report.Expiring) != 0 || report.Reloaded || engine.reloads != 0 {
		t.Fatalf("expected no reload, got %+v", report)
	}
}

func TestCheckOnceWithoutEngine(t *testing.T) {
	store := &fakeStore{certs: []models.Cert{{Hostname: "mail.example.com", ExpiresAt: time.Now().Add(time.Hour)}}}
	watcher := NewWatcher(store, nil, time.Hour, 30*24*time.Hour)
	report, err := watcher.CheckOnce(context.Background())
	if err != nil {
		t.Fatalf("check once: %v", err)
	}
	if len(report.Expiring) != 1 || report.Reloaded {
		t.Fatalf("expected report without reload, got %+v", report)
	}
}

func TestCheckOnceSurfacesReloadError(t *testing.T) {
	store := &fakeStore{certs: []models.Cert{{Hostname: "mail.example.com", ExpiresAt: time.Now().Add(time.Hour)}}}
	engine := &fakeEngine{err: errors.New("connection refused")}
	watcher := NewWatcher(store, engine, time.Hour, 30*24*time.Hour)
	report, err := watcher.CheckOnce(context.Background())
	if err != nil {
		t.Fatalf("reload failure should surface in report, got %v", err)
	}
	if report.Reloaded || report.ReloadError == "" {
		t.Fatalf("expected reload error in report, got %+v", report)
	}
}
