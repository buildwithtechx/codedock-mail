# Live-domain proof runbook

Proves codedock-mail against a real domain: DNS published, checks green,
send/receive round-trip. Ticks the deliverability checklist for real.

Driver script: `deploy/live_proof.py` (Python 3, standard library only).
It never publishes DNS for you; registrars differ, so that step stays manual.

## 0. Prerequisites

- A VPS with a public IPv4 address, ports `25, 465, 587, 143, 993, 4190`
  reachable inbound (see `deploy/README.md`), outbound port 25 unblocked
  by the provider, and PTR/rDNS set to `mail.<domain>`.
- A domain whose DNS you control.
- The stack up on the VPS (`docker compose up -d`), `maild` reachable,
  `MAILD_API_TOKEN` and `STALWART_SECRET` set.

## 1. Setup: domain, DKIM, proof mailbox

```bash
python3 deploy/live_proof.py setup --maild http://<host>:8081 \
  --token "$MAILD_API_TOKEN" --domain example.com
```

This ensures the domain and the `mail` DKIM selector exist, prints the DNS
records to publish, and creates `proof@example.com` (1 GB quota). The
mailbox password prints once; store it, the round-trip needs it.

## 2. Publish DNS, then poll checks

Publish every record from step 1 at your registrar, plus:

- `A mail.<domain> -> <vps-ip>`
- `PTR <vps-ip> -> mail.<domain>` (set at the VPS provider, not the registrar)

Then poll until green:

```bash
python3 deploy/live_proof.py checks --maild http://<host>:8081 \
  --token "$MAILD_API_TOKEN" --domain example.com --ip <vps-ip>
```

Green means MX, SPF, DMARC, DKIM, port-25 egress, and PTR all pass.
DNS propagation can take minutes; the poller retries 10 times by default.

## 3. Round-trip: send and receive

```bash
python3 deploy/live_proof.py roundtrip --host mail.example.com \
  --user proof@example.com
```

Sends a unique token message over submission (587/STARTTLS) and waits up
to 3 minutes for it over IMAP (993). Success prints `received ...`.

## 4. Deliverability checklist sign-off

- [ ] `checks` green (MX, SPF, DKIM, DMARC, port25, PTR).
- [ ] Round-trip received.
- [ ] `mail.<domain>` serves a valid LE cert (Stalwart ACME Automatic;
  `GET /api/certs` tracks expiry, watcher reloads near expiry).
- [ ] Thunderbird autoconfig and Outlook autodiscover answer for the domain.
- [ ] A backup exists (`POST /api/backups`) with both db and mail archives.

Record the domain, date, and any provider-specific notes (port-25 unblock
ticket, PTR panel path) in the release notes for the next install.
