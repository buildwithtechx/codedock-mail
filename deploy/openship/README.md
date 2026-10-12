# codedock-mail on Openship

`codedock-mail.json` is a starting-point catalog template for Openship
maintainers, written against the catalog schema (same shape as the `webmail`
template: `docker-compose` framework, services, generated secrets,
install-step settings, domain-only HTTP endpoints).

## Scope

The template wires **maild + Postgres + SnappyMail** against any IMAP/SMTP
backend. Engine provisioning (Stalwart principals, DKIM) works when
`STALWART_URL` points at a reachable v0.15.x engine; mail TCP ports cannot go
through the HTTP edge, so the engine itself stays a host-level install
exactly like Openship's own mail engine.

## Notes for integration

- `verified` is `false` and the entry is `unlisted` until an Openship
  maintainer validates it against the current catalog installer.
- Secrets use the catalog `generate: secret` convention (`MAILD_API_TOKEN`,
  `POSTGRES_PASSWORD`); the engine secret is pasted, never generated.
- `maild` is `exposed: false` — only the webmail service takes a public
  hostname, matching the no-public-admin-surface rule.
- SnappyMail domain entries (IMAP/SMTP hosts per domain) are added in its
  `/?admin` panel after install; see `apps/webmail-ext/DECISION.md`.
