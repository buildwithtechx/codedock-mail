# codedock-mail TODO

Standalone mail product. Own repo, own release. Installable in Codedock via HTTP seam. No shared DB.

## Decisions

- Backend: Go (`maild` daemon, Echo REST + OpenAPI, pgx Postgres, Docker SDK, cron).
- Reason: matches Codedock stack, single binary, no second runtime on mail host. Systems work (Docker, SSH, firewall, certs, DNS probes) reuses Codedock patterns.
- Admin UI: React + TypeScript, embeddable in Codedock dashboard.
- Webmail: reuse, never build from scratch (SnappyMail / Roundcube / Zero-fork as container).
- Integration: Codedock holds only a reference row (`server_id`, `domain`, `maild_url`, `webmail_url`). All mailbox writes go through `maild` HTTP API (token/mTLS). No FK across products.
- SMTP/IMAP (25/465/587/143/993/4190) bypass Traefik, raw TCP to host. Only webmail HTTPS goes through the edge.

## Engine choice (decided: Stalwart)

- Stalwart (single Rust binary: SMTP+IMAP+JMAP+Sieve, native Postgres, built-in SPF/DKIM/ARC/DMARC) behind the `Engine` interface (`Provision`, `AddMailbox`, `Health`) in `internal/provisioner`, so the backend stays swappable.
- [ ] Implement `StalwartEngine` against the Stalwart admin API (currently a stub returning not-implemented).

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
- [ ] Phase 2: DKIM/SPF/DMARC generation + DNS instructions, LE certs + cron renewal, port-25 / PTR / rDNS checks.
- [ ] Phase 3: webmail container, standalone compose, Codedock one-click template + dashboard page.
- [ ] Phase 4: backups (maildir + db), quota, multi-domain, sieve filters, audit log.

## Deliverability checklist

- [ ] MX, SPF, DKIM, DMARC, `autodiscover`/`autoconfig`, MTA-STS/TLS-RPT.
- [ ] PTR/rDNS matches `mail.<domain>`.
- [ ] Port 25 egress check + provider unblock runbook (Hetzner etc.).
- [ ] LE cert issue/renew for `mail.<domain>` + webmail host.
- [ ] Firewall: open 25/465/587/143/993/4190 TCP to host, webmail HTTP only via edge.

## Open questions

- [ ] Stalwart vs iRedMail reuse for v1?
- [ ] v1 scope: single-server/single-domain or multi-tenant from day one?
- [ ] Webmail pick: SnappyMail vs Roundcube vs Zero-fork?
