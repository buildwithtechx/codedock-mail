# Stalwart JMAP watch

The engine client targets the REST management API, which ends at Stalwart
v0.15.x. v0.16+ moves management to JMAP. The image stays pinned to
`v0.15.5` until a JMAP transport exists. This file is the watch procedure;
the monthly `stalwart-watch` workflow fails when upstream moves, which is
the signal to start the migration, not to bump the pin.

## What to check when the watch fires

1. Read the upstream release notes for the management API surface:
   which JMAP methods replace `POST /api/principal`, `/api/dkim`,
   `/api/dns/records/*`, and `/api/reload/certificate`.
2. Check whether v0.15.x still receives security fixes. If yes, staying
   pinned remains acceptable and the migration stays scheduled, not urgent.
3. Check the `deploy/stalwart.toml` keys against the new version's
   generated config; listener and storage keys drift between lines.

## Migration shape (planned, not started)

- New `internal/provisioner/jmap.go` implementing the existing `Engine`
  interface (`Provision`, `AddMailbox`, `UpdateMailboxQuota`,
  `CreateDKIM`, `DNSRecords`, `ReloadCertificates`, `Health`) over a
  JMAP session, plus Sieve stays on ManageSieve (port 4190).
- `NewStalwart` gains a transport switch (REST vs JMAP) selected by an
  `STALWART_MANAGEMENT=jmap|rest` variable, defaulting to whatever the
  pinned image speaks.
- `stalwart_test.go` patterns carry over: scripted HTTP server asserting
  JMAP request/response shapes per method.
- Unpin only after the live-domain proof passes against the new line.

## Manual check

```bash
python3 -c "import json,urllib.request; print(json.load(urllib.request.urlopen('https://api.github.com/repos/stalwartlabs/stalwart/releases/latest'))['tag_name'])"
```
