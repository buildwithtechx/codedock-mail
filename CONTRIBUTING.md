# Contributing to codedock-mail

Standalone mail product. Own repo, own release. Installable in Codedock via HTTP seam.

## Prerequisites

- **Go** `1.26+`
- **Node.js** `22+`
- **Docker** `24+`

## Setup

```bash
git clone https://github.com/buildwithtechx/codedock-mail.git
cd codedock-mail
cp .env.example .env
```

| Variable           | Default                                          | Description        |
| ------------------ | ------------------------------------------------ | ------------------ |
| `PORT`             | `8081`                                           | `maild` HTTP port  |
| `DATABASE_URL`     | `postgres://mail:mail@localhost:5432/mail?...`   | `maild` database   |
| `MAIL_DOMAIN`      | `example.com`                                    | Default mail domain|
| `MAIL_TEST_PG_URL` | —                                                | Store test database|

Migrations run automatically on `maild` startup. Postgres-backed store tests skip without `MAIL_TEST_PG_URL`:

```bash
MAIL_TEST_PG_URL=postgres://mail:mail@localhost:5432/mail_test?sslmode=disable go test ./...
```

## Start

```bash
go run ./cmd/maild        # API at http://localhost:8081
npm run dev:admin          # admin UI at http://localhost:3001
npm run dev:docs           # docs site
```

## Pull Requests

- **Target `main`.**
- **One PR = one thing.**
- **Title**: `type(scope): description` — e.g. `feat(store): add mailbox quota`.
- You must understand every line you submit. AI submissions you can't explain will be rejected.
- Merge requires `go vet`, `go test`, Biome format, and `tsc --noEmit` clean.

## Conventions

- Go: `internal/models` → `store` → `provisioner`/`dns`/`certs` → `api`. Narrow interfaces, no global state, wire up in `cmd/maild/main.go`.
- **No comments or GoDoc allowed.** JSON tags on every exported struct field.
- Frontend: named exports, `kebab-case` files, Biome formatter — never Prettier.
- Run `npm run fmt` before committing.
