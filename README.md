# codedock-mail

Standalone mail product. Own repo, own release. Installable in Codedock via HTTP seam. No shared DB.

`maild` (Go) owns domains, mailboxes, aliases, DKIM, certs. Codedock holds only a reference row and calls `maild` over HTTP. SMTP/IMAP bypass the edge as raw TCP; only webmail HTTPS goes through it.

See `TODO.md` for the build order and `docs/` for specs and runbooks.

## Support tiers

- **Tier 1 (supported):** Codedock-managed installs (one-click template,
  `/mail` dashboard page) and plain-compose VPS installs
  (`docker/docker-compose.yml`, see `deploy/README.md`).
- **Tier 2 (best-effort):** anything else that speaks compose (Coolify,
  Dokploy, and the like) using the same root compose file. No
  platform-specific code ships for these; the file is the contract.

## Layout

```sh
cmd/maild            # Go daemon entrypoint
internal/            # models, store, provisioner, dns, certs, api
apps/admin           # React admin, served by maild at / (token login)
apps/docs            # docs site
apps/webmail-ext     # webmail wrapper (FROM upstream + branding, plugins, defaults)
deploy/              # standalone compose install
docs/                # specs + runbooks
```

## Local development

Requirements: Go 1.26+, Node.js 22+, Docker.

```bash
cp .env.example .env
docker compose -f docker/docker-compose.yml up postgres -d
npm run dev
```

`npm run dev` runs the daemon under air (`:8081`) and the admin UI
under Vite (`:3001`, proxying `/api`) together. Run each alone with
`npm run dev:daemon` or `npm run dev:admin`. `maild` loads `.env` on
boot (explicit environment wins); air rebuilds when `.env` changes.
The default `DATABASE_URL` matches the compose `postgres` service, so
no extra config is needed for local dev.
