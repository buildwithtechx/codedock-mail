# codedock-mail on Dokku

Dokku deploys one container per app, so the stack becomes four apps plus a
linked Postgres. TCP mail ports are published directly; only `http`/`https`
schemes go through the nginx proxy.

```bash
dokku apps:create maild mail-engine webmail
dokku postgres:create maildb
dokku postgres:link maildb maild

dokku config:set maild \
  STALWART_URL=http://mail-engine:8080 \
  STALWART_USER=admin \
  STALWART_SECRET=<generate-me> \
  MAILD_API_TOKEN=<generate-me>
git push dokku main:master

dokku builder-dockerfile:set mail-engine dockerfile-path deploy/dokku/Dockerfile.engine
printf '%s' "$(cat deploy/stalwart.toml)" > /var/lib/dokku/data/storage/mail-engine/config.toml
dokku storage:ensure-directory mail-engine
dokku storage:mount mail-engine /var/lib/dokku/data/storage/mail-engine/config.toml:/opt/stalwart/etc/config.toml
dokku storage:mount mail-engine /var/lib/dokku/data/storage/mail-engine/data:/opt/stalwart/data
dokku config:set mail-engine DOKKU_DOCKERFILE_START_CMD=""
git push dokku main:master --app mail-engine

dokku git:from-image webmail djmaze/snappymail:2.38.2
dokku storage:ensure-directory webmail
dokku storage:mount webmail /var/lib/dokku/data/storage/webmail:/var/lib/snappymail
dokku domains:add webmail mail.example.com
dokku letsencrypt:enable webmail
```

Then publish the engine's mail ports and lock `maild` down to the tailnet:

```bash
dokku ports:add mail-engine 25:25 465:465 587:587 143:143 993:993 4190:4190
dokku ports:remove maild http:80:8081 https:443:8081
```

Change the fallback-admin secret inside the mounted `config.toml` before first
boot, open firewall TCP `25, 465, 587, 143, 993, 4190`, confirm outbound port
25 is unblocked, and set PTR/rDNS to `mail.<domain>`. Finish in the Stalwart
webadmin (ACME provider, per-domain Automatic) and SnappyMail `/?admin`
(domain IMAP/SMTP entries); `maild` owns domains, DKIM, DNS, and checks after
that.
