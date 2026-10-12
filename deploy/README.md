# deploy

`docker-compose.yml` at the repo root is the stack: `maild` + Postgres +
engine + webmail. Point any compose-based setup at the repo and it works —
plain Docker, Coolify, Dokploy, and the like.

```bash
cp .env.example .env
$EDITOR deploy/stalwart.toml .env
docker compose up -d
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
   `GET /api/domains/:id/checks`.

## Different mechanism

- Dokku (`deploy/dokku/`, removed): its nginx proxy does not do plain TCP, so
  mail ports would need proxy-disable plus manual docker flags with no
  zero-downtime deploys, and inter-app hostnames follow its own convention.
  None of that was verifiable here. On a Dokku host, run the compose file
  directly instead — Dokku is Docker underneath and the mail ports don't clash
  with its proxy.
