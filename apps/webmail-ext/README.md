# webmail-ext

Thin wrapper around the reused upstream webmail container. Upstream owns rendering and sending; this owns look, login, and defaults.

- `Dockerfile` — `FROM` upstream plus branding, plugins, default config.
- `branding/` — logo, colors, login page.
- `config/` — IMAP/SMTP defaults (`mail.<domain>`), ports, autoconfig.
- `plugins/` — pinned webmail plugins (signatures, filters, vacation replies).
