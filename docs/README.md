# docs

Specs and runbooks. The docs site lives in `apps/docs`; this folder holds the source material.

- Engine decision (Stalwart vs iRedMail reuse)
- DNS checklist (MX, SPF, DKIM, DMARC, autodiscover, MTA-STS)
- Port-25 runbook (egress check, provider unblock, PTR/rDNS)
- API spec (`internal/api`, OpenAPI)

## API

`maild` listens on `PORT` (default `8081`). Without `STALWART_URL` it runs
local-only: domains, mailboxes, and DKIM metadata are stored, but nothing is
provisioned on an engine and mailbox creation returns no password.

| Method   | Path                         | Notes                                              |
| -------- | ---------------------------- | -------------------------------------------------- |
| GET      | `/health`                    | liveness                                           |
| GET/POST | `/api/domains`               | engine provisioned first when configured           |
| GET/POST | `/api/domains/:id/mailboxes` | password returned once on create, never stored     |
| POST     | `/api/domains/:id/dkim`      | engine generates the key, metadata stored          |
| GET      | `/api/domains/:id/dns`       | local MX/SPF/DMARC/DKIM plus live engine records   |
| GET      | `/api/domains/:id/checks`    | MX/SPF/DMARC/DKIM, port-25 egress, `?ip=` adds PTR |
| GET      | `/api/domains/:id`           | single domain |
| DELETE   | `/api/domains/:id`           | engine principal removed first, then local row (cascade) |
| DELETE   | `/api/domains/:domainId/mailboxes/:mailboxId` | engine principal removed first |
| GET      | `/api/certs`                 | tracked certificate expiries |
| POST     | `/api/certs/probe`           | TLS-handshake expiry probe (implicit TLS ports), upserts tracking row |
| GET      | `/autoconfig/mail/config-v1.1.xml?emailaddress=` | Thunderbird client config (also under `/.well-known/`) |
| GET/POST | `/autodiscover/autodiscover.xml` | Outlook 2006a POX (XML body or `?Email=`) |

## Certs

Issuance stays engine-side: enable an ACME provider in Stalwart and set each
domain to Automatic. `maild` tracks expiry with a TLS probe (no engine
dependency), warns below 30 days, and calls `GET /api/reload/certificate`
hourly when anything is near expiry.

## Engine

Stalwart `v0.15.5` (pinned: last line with the REST management API) over HTTP
Basic (`STALWART_USER`, `STALWART_SECRET`). Principals are addressed by name.
Secrets are bcrypt hashes (`$2a$...`), which Stalwart verifies natively. DKIM
keys are generated engine-side; `maild` keeps only the selector, public key,
and DNS record for instructions. Duplicate creates surface
`fieldAlreadyExists` in a 200 envelope, mapped to 409.
