# codedock-mail TODO

Standalone mail product. Own repo, own release. Installable in Codedock via HTTP seam. No shared DB.

## Decisions

- Backend: Go (`maild` daemon, Echo REST + OpenAPI, pgx Postgres, Docker SDK, cron).
- Reason: matches Codedock stack, single binary, no second runtime on mail host. Systems work (Docker, SSH, firewall, certs, DNS probes) reuses Codedock patterns.
- Admin UI: React + TypeScript, embeddable in Codedock dashboard.
- Webmail: reuse, never build from scratch (SnappyMail / Roundcube / Zero-fork as container).
- Integration: Codedock holds only a reference row (`server_id`, `domain`, `maild_url`, `webmail_url`). All mailbox writes go through `maild` HTTP API (token/mTLS). No FK across products.
- SMTP/IMAP (25/465/587/143/993/4190) bypass Traefik, raw TCP to host. Only webmail HTTPS goes through the edge.

## Engine choice (decided: Stalwart v0.15.5, REST management)

- Stalwart (single Rust binary: SMTP+IMAP+JMAP+Sieve, native Postgres, built-in SPF/DKIM/ARC/DMARC) behind the `Engine` interface (`Provision`, `AddMailbox`, `Health`) in `internal/provisioner`, so the backend stays swappable.
- Pinned `stalwartlabs/stalwart:v0.15.5`: last line with the REST management API the client targets. v0.16+ moves management to JMAP — tracked below, not implemented.
- `deploy/stalwart.toml` mirrors the image's own generated config (verified by boot: layout `/opt/stalwart/{etc,data,logs}`, plaintext fallback-admin accepted, all listener keys exact).
- Verified live against v0.15.5: auth, domain/user create, `description` rejected (dropped from payloads), DKIM POST returns null (metadata recovered from DNS records endpoint), duplicate → `fieldAlreadyExists` envelope (maps to 409), reload, deletes.
- [ ] Track Stalwart v0.16 JMAP management migration; client will need a JMAP transport before unpinning.
- [x] Engine client: Stalwart REST (`Provision`, `RemoveDomain`, `AddMailbox`, `RemoveMailbox`, `CreateDKIM`, `DNSRecords`, `ReloadCertificates`, `Health`), bcrypt-hashed secrets, mailbox password returned once at creation and never stored.
- [x] DKIM metadata store (`002_dkim_metadata`) + `POST /api/domains/:id/dkim`, DNS instructions via `GET /api/domains/:id/dns` (local MX/SPF/DMARC/DKIM + live engine records).
- [x] Deliverability checks via `GET /api/domains/:id/checks` (MX/SPF/DMARC/DKIM, port-25 egress, optional PTR via `?ip=`).
- [ ] LE certs: engine-side ACME config + `certs` expiry tracking.

## Repo layout

```text
codedock-mail/
  cmd/maild/main.go
  internal/models/      domain.go, mailbox.go, dkim.go, cert.go
  internal/store/       postgres queries (pgx)
  internal/provisioner/ engine.go, webmail.go, firewall.go, ports.go
  internal/dns/         spf.go, dkim.go, dmarc.go, autodiscover.go
  internal/certs/       acme.go, renew.go
  internal/api/         routes.go, openapi.json
  apps/admin/           React admin (embeddable, standalone dev)
  apps/docs/            docs site (install, DNS, deliverability, API)
  apps/webmail-ext/     webmail wrapper (Dockerfile FROM upstream + branding, plugins, defaults, auth bridge)
  deploy/compose.yml    standalone install
  docs/                 specs + runbooks (DNS checklist, port-25, engine decision)
```

## Codedock adapter (thin, lives in codedock repo, not here)

- [ ] `internal/services/mail/` HTTP client to `maild`.
- [ ] `internal/handlers/mail.go` proxy + reference row.
- [ ] Dashboard route `/mail` embedding `apps/admin` build.
- [ ] One-click template `mail.json` installing the `maild` stack.

## MVP phases

- [x] Phase 1: `maild` + Postgres. One domain, mailbox CRUD via API (`internal/store` migration `001_initial`, `internal/api` routes), health endpoint.
- [x] Phase 2: DKIM/SPF/DMARC generation + DNS instructions, LE cert expiry tracking + hourly watcher with engine reload, port-25 / PTR / rDNS checks. Issuance stays engine-side (Stalwart ACME provider + per-domain Automatic).
- [x] Phase 3a (this repo): full lifecycle (`GET/DELETE` domains and mailboxes), Thunderbird + Outlook autoconfig, `MAILD_API_TOKEN` auth, maild Dockerfile, webmail default (SnappyMail 2.38.2, see `apps/webmail-ext/DECISION.md`), adapter contract (`docs/ADAPTER.md`), live Postgres proof.
- [x] Phase 3b: engine config verified by boot (`deploy/stalwart.toml`), SnappyMail pinned with admin-panel wiring documented, install assets for Coolify/Dokku/Dokploy/Openship (`deploy/*`).
- [ ] Phase 3c (codedock repo): one-click template + dashboard page per `docs/ADAPTER.md`.

## Install matrix

| Platform | Asset | Notes |
| -------- | ----- | ----- |
| Any VPS / Coolify / Dokploy | `docker-compose.yml` + `deploy/stalwart.toml` | single shared stack file, see `deploy/README.md` |
| Dokku | `deploy/dokku/` | four apps + linked Postgres, TCP ports published directly |
| Openship | `deploy/openship/codedock-mail.json` | unverified starting-point catalog template |
- [ ] Phase 4: backups (maildir + db), quota, multi-domain, sieve filters, audit log.

## Deliverability checklist

- [ ] MX, SPF, DKIM, DMARC, `autodiscover`/`autoconfig`, MTA-STS/TLS-RPT.
- [ ] PTR/rDNS matches `mail.<domain>`.
- [ ] Port 25 egress check + provider unblock runbook (Hetzner etc.).
- [ ] LE cert issue/renew for `mail.<domain>` + webmail host.
- [ ] Firewall: open 25/465/587/143/993/4190 TCP to host, webmail HTTP only via edge.

## Open questions

- [ ] v1 scope: single-server/single-domain or multi-tenant from day one?
- [ ] Stalwart v0.16 JMAP management migration (client needed before unpinning v0.15.5).
