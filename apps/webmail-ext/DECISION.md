# Webmail decision

Default: **SnappyMail** (`apps/webmail-ext/Dockerfile` builds `FROM` it, overridable via `UPSTREAM_IMAGE`).

| Option | For | Against |
| ------ | --- | ------- |
| SnappyMail | Small single image, IMAP/SMTP-native, no OAuth baggage, active RainLoop fork | Branding and defaults mapped per release, PHP internals |
| Roundcube | Mature, plugin ecosystem | Heavier (web server + PHP + prefs DB), dated UI |
| Zero-fork | Modern React UI, env-driven config with a known contract | Tracks a fork, second JS runtime on the host |

`branding/` holds the login overlay (logo, theme) applied over the pinned
upstream image. `config/` holds the deploy-time defaults (IMAP/SMTP hosts and
ports following the `mail.<domain>` convention). Both are mapped against the
pinned upstream version at integration time, never against `latest` blindly.
