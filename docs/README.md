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

| Method | Path | Notes |
| ------ | ---- | ----- |
| GET | `/health` | liveness |
| GET/POST | `/api/domains` | engine provisioned first when configured |
| GET/POST | `/api/domains/:id/mailboxes` | password returned once on create, never stored |
| POST | `/api/domains/:id/dkim` | engine generates the key, metadata stored |
| GET | `/api/domains/:id/dns` | local MX/SPF/DMARC/DKIM plus live engine records |
| GET | `/api/domains/:id/checks` | MX/SPF/DMARC/DKIM, port-25 egress, `?ip=` adds PTR |

## Engine

Stalwart REST management API over HTTP Basic (`STALWART_USER`, `STALWART_SECRET`).
Principals are addressed by name. Secrets are bcrypt hashes (`$2a$...`), which
Stalwart verifies natively. DKIM keys are generated engine-side; `maild` keeps
only the selector, public key, and DNS record for instructions.

