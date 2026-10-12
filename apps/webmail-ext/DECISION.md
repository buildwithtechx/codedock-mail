# Webmail decision

Default: **SnappyMail 2.38.2** (`djmaze/snappymail:2.38.2`, verified latest
upstream publish). `apps/webmail-ext/Dockerfile` builds `FROM` it, overridable
via `UPSTREAM_IMAGE`. The stock compose service runs upstream directly; build
the overlay for branded distributions.

| Option | For | Against |
| ------ | --- | ------- |
| SnappyMail | Small single image, IMAP/SMTP-native, no OAuth baggage, active RainLoop fork | Branding and defaults mapped per release, PHP internals |
| Roundcube | Mature, plugin ecosystem | Heavier (web server + PHP + prefs DB), dated UI |
| Zero-fork | Modern React UI, env-driven config with a known contract | Tracks a fork, second JS runtime on the host |

`branding/` holds the login overlay (logo, theme) applied over the pinned
upstream image. `config/` holds the deploy-time defaults (IMAP/SMTP hosts and
ports following the `mail.<domain>` convention). Both are mapped against the
pinned upstream version at integration time, never against `latest` blindly.

SnappyMail specifics (verified against upstream docs): the image serves on
`8888`; all state (domain entries, branding, sessions) lives in the
`/var/lib/snappymail` volume. Domain IMAP/SMTP entries are created in the
admin panel at `/?admin` (initial password in
`/var/lib/snappymail/_data_/_default_/admin_password.txt`); branding (title,
loading text, favicon, logo) lives under the same Branding menu. No database
is required.
