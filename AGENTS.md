# Agent Instructions

## Code Style

- **Max 350 lines per file.** If a file exceeds this, extract helpers, adapters, or types into separate files.
- **One component per file.** Never cram thousands of lines into a single file. Break components down into individual files.
- **Name files in `kebab-case`** (e.g. `project-card.tsx`, `use-logs-stream.ts`).
- Use named exports over default exports.
- **No comments or GoDoc/JSDoc allowed.** Code should be self-explanatory. If logic is truly non-obvious, refactor to make it clear rather than adding a comment.

## Workflow

- **Do not run build or test commands after every change.** Just make the code change. If something breaks, the user will say so.
- Run `npm run fmt` (`go fmt ./...` and `npm run format:fix`) before committing or finishing a session. NEVER run `prettier` (`npx prettier`) — Biome is our strict formatter.
- Prefer `read`/`grep`/`glob` tools over `bash` for file exploration.
- When making edits, read the file first, then use `edit` for targeted changes.

## Stack

| Layer             | Tech                                                       |
| ----------------- | ---------------------------------------------------------- |
| Backend (`maild`) | Go (`cmd`, `internal/`), Echo, `pgx/v5`, Docker SDK, cron  |
| Admin UI          | React 19, TanStack Router, TanStack Query, Tailwind CSS v4 |
| Docs site         | Astro + Starlight (`apps/docs`)                            |
| Webmail           | Reused upstream container, wrapped in `apps/webmail-ext`   |

## Conventions

- **Go layers (`internal/`):** `models/` domain structs and DTOs, `store/` Postgres persistence, `provisioner/` engine and webmail lifecycle, `dns/` records, `certs/` ACME and renewal, `api/` HTTP routes.
- **Consumer-Defined Interfaces:** Define narrow interfaces where consumed (`Accept interfaces, return structs`).
- **File naming:** lowercase snake_case for Go (`container_health.go`), `kebab-case` for TS.
- **Package naming:** short, lowercase, single word.
- **Error handling:** always check errors; wrap with `fmt.Errorf("context: %w", err)`.
- **No global state.** Pass dependencies via struct fields, wire up in `cmd/maild/main.go`.
- **JSON tags** on every exported struct field.
- Avoid `init()` functions. Use explicit constructor functions instead.
- Unit tests stay co-located with source files.

## Apps & Workspaces

```sh
npm run dev:admin    # admin UI dev server
npm run dev:docs     # docs site dev server
npm run build:daemon # builds bin/maild
npm run fmt          # go fmt + Biome
```
