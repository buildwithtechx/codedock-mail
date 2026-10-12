# codedock-mail

Standalone mail product. Own repo, own release. Installable in Codedock via HTTP seam. No shared DB.

`maild` (Go) owns domains, mailboxes, aliases, DKIM, certs. Codedock holds only a reference row and calls `maild` over HTTP. SMTP/IMAP bypass the edge as raw TCP; only webmail HTTPS goes through it.

See `TODO.md` for the build order and `docs/` for specs and runbooks.

## Layout

```sh
cmd/maild            # Go daemon entrypoint
internal/            # models, store, provisioner, dns, certs, api
apps/admin           # React admin, embeddable in Codedock dashboard
apps/docs            # docs site
apps/webmail-ext     # webmail wrapper (FROM upstream + branding, plugins, defaults)
deploy/              # standalone compose install
docs/                # specs + runbooks
```

## Local development

Requirements: Go 1.26+, Node.js 22+, Docker.

```bash
cp .env.example .env
go run ./cmd/maild
npm run dev:admin
```
