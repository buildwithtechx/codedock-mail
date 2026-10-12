# Security Policy

codedock-mail manages mailbox credentials, DKIM private keys, TLS certificates, and talks to the Docker daemon. Security issues here can compromise every mailbox on a host.

## Supported Versions

| Version | Supported |
| ------- | --------- |
| `0.1.x` | ✅        |

## Reporting a Vulnerability

**Do not report vulnerabilities through public issues, PRs, or discussions.** Instead:

- Open a private [GitHub Security Advisory](https://github.com/buildwithtechx/codedock-mail/security/advisories/new), or
- Email **security@codedock.run**.

### What to Include

- Type of issue (e.g. auth bypass, secret leakage, SQL injection, container escape).
- Full paths of the source file(s) involved.
- Step-by-step reproduction or proof-of-concept.

We commit to acknowledging disclosures within **48 hours**.
