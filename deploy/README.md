# deploy

`docker/docker-compose.yml` is the stack: `maild` + Postgres +
engine + webmail. Point any compose-based setup at the repo and it works —
plain Docker, Coolify, Dokploy, and the like.

```bash
cp .env.example .env
$EDITOR deploy/stalwart.toml .env
docker compose -f docker/docker-compose.yml up -d
```

`stalwart.toml` is mounted read-only into the engine; the data volume holds
everything else. If the engine ignores the file (wrong mount path for the
image revision), it prints generated admin credentials in `docker logs` —
copy those into `.env` as `STALWART_SECRET` instead.

## After boot

1. Open firewall TCP `25, 465, 587, 143, 993, 110, 995, 4190`. Confirm
   outbound port 25 is unblocked and set PTR/rDNS to `mail.<domain>`.
2. In the Stalwart webadmin enable an ACME provider and set each domain to
   Automatic. In SnappyMail (`/?admin`) add the domain IMAP/SMTP entries.
3. Drive everything after that through `maild`: create the domain and DKIM
   key, publish the DNS records it returns, and watch
   `GET /api/domains/:id/checks`. For a first real domain, follow
   `docs/LIVE_PROOF.md` (`scripts/live-proof.py`).
4. Backups land in the `mail_backups` volume: a Postgres dump plus a
   tarball of the engine data (`MAIL_DATA_DIR`). The db restores via
   `POST /api/backups/:id/restore`; the mail archive restores via
   `POST /api/backups/:id/restore-mail`, which refuses while the
   engine is up and snapshots the current data dir first — so stop
   the engine, restore, start it again. Set `BACKUP_SCHEDULE` (cron,
   e.g. `0 3 * * *`) for unattended runs; `BACKUP_RETENTION` keeps
   that many newest backups (`0` keeps all).

## Different mechanism

- Dokku (`deploy/dokku/`, removed): its nginx proxy does not do plain TCP, so
  mail ports would need proxy-disable plus manual docker flags with no
  zero-downtime deploys, and inter-app hostnames follow its own convention.
  None of that was verifiable here. On a Dokku host, run the compose file
  directly instead — Dokku is Docker underneath and the mail ports don't clash
  with its proxy.
