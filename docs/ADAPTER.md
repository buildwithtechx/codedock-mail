# Codedock adapter contract

How the Codedock control plane installs and talks to `codedock-mail`.
Implemented in the codedock repo; this file is the contract it builds against.

## Install

Mail ships as a compose stack (see `docker/docker-compose.yml`): `maild` + Postgres +
engine + webmail. Codedock deploys it like any one-click app and stores one
reference row per install:

| Field | Meaning |
| ----- | ------- |
| `server_id` | target host |
| `domain` | primary mail domain |
| `maild_url` | `http://<host>:8081` (internal, never public) |
| `webmail_url` | public HTTPS URL of the webmail service |
| `api_token` | `MAILD_API_TOKEN` value, kept server-side only |

No shared database. No foreign keys across products. Uninstall deletes the
stack and the reference row.

## Auth

`Authorization: Bearer <api_token>` on every `/api/*` route. Public without
a token: `/health`, `/autoconfig/*`, `/autodiscover/*`, `/.well-known/*`.

## Lifecycle mapping

| Codedock action | maild calls |
| --------------- | ----------- |
| Install | deploy stack, `POST /api/domains`, `POST /api/domains/:id/dkim`, `GET /api/domains/:id/dns` for the DNS instructions page |
| Add mailbox | `POST /api/domains/:id/mailboxes` (show password once) |
| Remove mailbox | `DELETE /api/domains/:domainId/mailboxes/:mailboxId` |
| Health | `/health` plus `GET /api/domains/:id/checks` for the status page |
| Cert watch | `GET /api/certs` |
| Quota | `PUT /api/domains/:domainId/mailboxes/:mailboxId/quota` |
| Backup | `POST /api/backups`, `GET /api/backups` for the backup page |
| Audit | `GET /api/audit` |
| Uninstall | `DELETE /api/domains/:id`, delete stack |

Engine ordering inside `maild` is engine-first, store-second on create and
engine-first on delete, so a failed engine call never leaves a half-provisioned
local row behind.
