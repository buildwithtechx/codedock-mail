package store

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("MAIL_TEST_PG_URL")
	if dsn == "" {
		t.Skip("MAIL_TEST_PG_URL is not set")
	}
	ctx := context.Background()
	st, err := Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	if _, err := st.pool.Exec(ctx, "TRUNCATE domains, mailboxes, dkim_keys, certs CASCADE"); err != nil {
		t.Fatalf("truncate test db: %v", err)
	}
	return st
}

func TestCreateAndListDomains(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	created, err := st.CreateDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("create domain: %v", err)
	}
	if created.Name != "example.com" {
		t.Fatalf("expected example.com, got %q", created.Name)
	}
	if _, err := st.CreateDomain(ctx, "example.com"); !errors.Is(err, ErrExists) {
		t.Fatalf("duplicate domain should return ErrExists, got %v", err)
	}
	listed, err := st.ListDomains(ctx)
	if err != nil {
		t.Fatalf("list domains: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("expected one domain %s, got %+v", created.ID, listed)
	}
}

func TestDKIMKeys(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	d, err := st.CreateDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("create domain: %v", err)
	}
	created, err := st.CreateDKIMKey(ctx, models.DKIMKey{
		DomainID:  d.ID,
		Selector:  "mail",
		Algorithm: "Ed25519",
		PublicKey: "abc",
		DNSRecord: "v=DKIM1; k=ed25519; p=abc",
	})
	if err != nil {
		t.Fatalf("create dkim key: %v", err)
	}
	if created.Selector != "mail" || created.PublicKey != "abc" {
		t.Fatalf("unexpected key %+v", created)
	}
	if _, err := st.CreateDKIMKey(ctx, models.DKIMKey{DomainID: d.ID, Selector: "mail", Algorithm: "Ed25519"}); !errors.Is(err, ErrExists) {
		t.Fatalf("duplicate dkim selector should return ErrExists, got %v", err)
	}
	latest, err := st.GetDKIMKey(ctx, d.ID)
	if err != nil {
		t.Fatalf("get dkim key: %v", err)
	}
	if latest.ID != created.ID {
		t.Fatalf("expected latest key %s, got %s", created.ID, latest.ID)
	}
	if _, err := st.GetDKIMKey(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing domain should return ErrNotFound, got %v", err)
	}
}

func TestCreateMailboxBuildsAddress(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	d, err := st.CreateDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("create domain: %v", err)
	}
	m, err := st.CreateMailbox(ctx, d.ID, "alice", 0)
	if err != nil {
		t.Fatalf("create mailbox: %v", err)
	}
	if m.Address != "alice@example.com" {
		t.Fatalf("expected alice@example.com, got %q", m.Address)
	}
	if _, err := st.CreateMailbox(ctx, d.ID, "alice", 0); !errors.Is(err, ErrExists) {
		t.Fatalf("duplicate mailbox should return ErrExists, got %v", err)
	}
	if _, err := st.CreateMailbox(ctx, "missing", "bob", 0); !errors.Is(err, ErrNotFound) {
		t.Fatalf("mailbox on missing domain should return ErrNotFound, got %v", err)
	}
	listed, err := st.ListMailboxes(ctx, d.ID)
	if err != nil {
		t.Fatalf("list mailboxes: %v", err)
	}
	if len(listed) != 1 || listed[0].Address != "alice@example.com" {
		t.Fatalf("expected one mailbox alice@example.com, got %+v", listed)
	}
	quota, err := st.CreateMailbox(ctx, d.ID, "quota", 512)
	if err != nil {
		t.Fatalf("create mailbox with quota: %v", err)
	}
	if quota.QuotaMB != 512 {
		t.Fatalf("expected quota 512, got %d", quota.QuotaMB)
	}
	got, err := st.GetMailbox(ctx, quota.ID)
	if err != nil || got.QuotaMB != 512 {
		t.Fatalf("expected persisted quota 512, got %+v %v", got, err)
	}
}

func TestAuditLog(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	entry, err := st.Log(ctx, models.AuditEntry{
		Action:       "domain.create",
		ResourceType: "domain",
		ResourceID:   "d-1",
		Detail:       `{"name":"example.com"}`,
	})
	if err != nil {
		t.Fatalf("log audit entry: %v", err)
	}
	if entry.Actor != "operator" || entry.ID == "" {
		t.Fatalf("unexpected entry %+v", entry)
	}
	entries, err := st.ListAudit(ctx, 10)
	if err != nil || len(entries) != 1 || entries[0].Action != "domain.create" {
		t.Fatalf("expected one entry, got %+v %v", entries, err)
	}
}

func TestCertsUpsertAndExpiring(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	soon := time.Now().Add(time.Hour)
	later := time.Now().Add(90 * 24 * time.Hour)
	first, err := st.UpsertCert(ctx, "mail.example.com", soon)
	if err != nil {
		t.Fatalf("upsert cert: %v", err)
	}
	second, err := st.UpsertCert(ctx, "mail.example.com", later)
	if err != nil {
		t.Fatalf("upsert cert again: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected same cert row on hostname conflict, got %s then %s", first.ID, second.ID)
	}
	expiring, err := st.ListExpiring(ctx, time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("list expiring: %v", err)
	}
	if len(expiring) != 0 {
		t.Fatalf("expected no expiring certs, got %+v", expiring)
	}
	if _, err := st.UpsertCert(ctx, "old.example.com", time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("upsert expired cert: %v", err)
	}
	expiring, err = st.ListExpiring(ctx, time.Now())
	if err != nil {
		t.Fatalf("list expiring: %v", err)
	}
	if len(expiring) != 1 || expiring[0].Hostname != "old.example.com" {
		t.Fatalf("expected one expired cert, got %+v", expiring)
	}
}

func TestDomainByNameAndDelete(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	created, err := st.CreateDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("create domain: %v", err)
	}
	found, err := st.GetDomainByName(ctx, "example.com")
	if err != nil || found.ID != created.ID {
		t.Fatalf("expected domain %s, got %+v %v", created.ID, found, err)
	}
	if _, err := st.GetDomainByName(ctx, "missing.test"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := st.CreateMailbox(ctx, created.ID, "alice", 0); err != nil {
		t.Fatalf("create mailbox: %v", err)
	}
	if err := st.DeleteDomain(ctx, created.ID); err != nil {
		t.Fatalf("delete domain: %v", err)
	}
	if _, err := st.GetDomain(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	if _, err := st.ListMailboxes(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected mailboxes gone with domain, got %v", err)
	}
	if err := st.DeleteDomain(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound on second delete, got %v", err)
	}
}

func TestMailboxGetAndDelete(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	d, err := st.CreateDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("create domain: %v", err)
	}
	created, err := st.CreateMailbox(ctx, d.ID, "alice", 0)
	if err != nil {
		t.Fatalf("create mailbox: %v", err)
	}
	found, err := st.GetMailbox(ctx, created.ID)
	if err != nil || found.Address != "alice@example.com" {
		t.Fatalf("expected alice@example.com, got %+v %v", found, err)
	}
	updated, err := st.UpdateMailboxQuota(ctx, created.ID, 1024)
	if err != nil || updated.QuotaMB != 1024 || updated.Address != "alice@example.com" {
		t.Fatalf("expected quota 1024 on alice@example.com, got %+v %v", updated, err)
	}
	if _, err := st.UpdateMailboxQuota(ctx, "missing", 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := st.GetMailbox(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := st.DeleteMailbox(ctx, created.ID); err != nil {
		t.Fatalf("delete mailbox: %v", err)
	}
	if _, err := st.GetMailbox(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
