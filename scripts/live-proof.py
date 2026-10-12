#!/usr/bin/env python3
import argparse
import getpass
import imaplib
import json
import smtplib
import ssl
import sys
import time
import urllib.error
import urllib.request
from email.message import EmailMessage


def api_call(maild, token, method, path, body=None, extra_headers=None):
    data = None
    headers = {"Accept": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    if extra_headers:
        headers.update(extra_headers)
    if body is not None:
        data = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(maild.rstrip("/") + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            raw = resp.read().decode()
            return resp.status, json.loads(raw) if raw else None
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode()[:500]
        return exc.code, {"_error": detail}


def fail(message):
    print("live-proof: " + message, file=sys.stderr)
    raise SystemExit(1)


def cmd_setup(args):
    status, health = api_call(args.maild, args.token, "GET", "/health")
    if status != 200:
        fail("maild not healthy at %s (status %s)" % (args.maild, status))
    print("maild ok: %s" % health)

    status, domains = api_call(args.maild, args.token, "GET", "/api/domains")
    if status != 200:
        fail("list domains failed: %s" % domains)
    domain = next((d for d in domains if d["name"] == args.domain), None)
    if domain is None:
        status, domain = api_call(args.maild, args.token, "POST", "/api/domains", {"name": args.domain})
        if status == 409:
            fail("domain exists on engine but not locally; delete it in the engine and retry")
        if status != 201:
            fail("create domain failed: %s" % domain)
        print("domain created: %s" % domain["id"])
    else:
        print("domain exists: %s" % domain["id"])
    domain_id = domain["id"]

    status, dkim = api_call(args.maild, args.token, "POST", "/api/domains/%s/dkim" % domain_id, {"selector": "mail"})
    if status == 201:
        print("dkim created: selector=%s" % dkim["selector"])
    elif status == 409:
        print("dkim already exists, continuing")
    else:
        fail("create dkim failed: %s" % dkim)

    status, dns = api_call(args.maild, args.token, "GET", "/api/domains/%s/dns" % domain_id)
    if status != 200:
        fail("dns instructions failed: %s" % dns)
    print("\nPublish these DNS records, then run: checks")
    for record in dns["records"]:
        print("  %-4s %-32s %s  [%s]" % (record["type"], record["name"], record["value"], record["source"]))
    if dns.get("engineError"):
        print("  engine dns warning: %s" % dns["engineError"])

    status, boxes = api_call(args.maild, args.token, "GET", "/api/domains/%s/mailboxes" % domain_id)
    if status != 200:
        fail("list mailboxes failed: %s" % boxes)
    address = args.mailbox + "@" + args.domain
    if any(b["address"] == address for b in boxes):
        print("\nmailbox %s exists; its password is unknown to maild (shown once at creation)." % address)
        print("Delete and recreate it to get a fresh password for roundtrip.")
        return
    status, created = api_call(
        args.maild, args.token, "POST", "/api/domains/%s/mailboxes" % domain_id,
        {"localPart": args.mailbox, "quotaMb": 1024},
    )
    if status != 201:
        fail("create mailbox failed: %s" % created)
    print("\nmailbox %s created." % address)
    print("MAILBOX PASSWORD (shown once, store it now): %s" % created.get("password"))


def all_passing(checks):
    names = {c["name"] for c in checks}
    if not {"mx", "spf", "dmarc", "dkim"}.issubset(names):
        return False
    return all(c["passed"] for c in checks)


def cmd_checks(args):
    status, domains = api_call(args.maild, args.token, "GET", "/api/domains")
    if status != 200:
        fail("list domains failed: %s" % domains)
    domain = next((d for d in domains if d["name"] == args.domain), None)
    if domain is None:
        fail("domain %s not found; run setup first" % args.domain)
    path = "/api/domains/%s/checks" % domain["id"]
    if args.ip:
        path += "?ip=" + args.ip
    for attempt in range(1, args.tries + 1):
        status, report = api_call(args.maild, args.token, "GET", path)
        if status != 200:
            fail("checks failed: %s" % report)
        print("attempt %d/%d:" % (attempt, args.tries))
        for check in report["checks"]:
            print("  %s %-8s %s" % ("PASS" if check["passed"] else "FAIL", check["name"], check["detail"]))
        if all_passing(report["checks"]):
            print("all checks green")
            return
        if attempt < args.tries:
            time.sleep(args.wait)
    fail("checks still red after %d attempts" % args.tries)


def cmd_roundtrip(args):
    password = args.password or getpass.getpass("mailbox password: ")
    marker = "live-proof-%d" % int(time.time())
    msg = EmailMessage()
    msg["From"] = args.user
    msg["To"] = args.user
    msg["Subject"] = marker
    msg.set_content("codedock-mail live-domain proof %s" % marker)

    context = ssl.create_default_context()
    with smtplib.SMTP(args.host, args.smtp_port, timeout=30) as smtp:
        smtp.starttls(context=context)
        smtp.login(args.user, password)
        smtp.send_message(msg)
    print("sent %s via %s:%d" % (marker, args.host, args.smtp_port))

    deadline = time.time() + args.timeout
    while time.time() < deadline:
        with imaplib.IMAP4_SSL(args.host, args.imap_port, ssl_context=context) as imap:
            imap.login(args.user, password)
            imap.select("INBOX")
            _, data = imap.search(None, "SUBJECT", '"%s"' % marker)
            if data and data[0].strip():
                print("received %s over IMAP" % marker)
                return
        time.sleep(10)
    fail("message %s not received within %ds" % (marker, args.timeout))


def main():
    parser = argparse.ArgumentParser(description="codedock-mail live-domain proof driver")
    parser.add_argument("--maild", default="http://localhost:8081")
    parser.add_argument("--token", default="")
    sub = parser.add_subparsers(dest="command", required=True)

    setup = sub.add_parser("setup")
    setup.add_argument("--domain", required=True)
    setup.add_argument("--mailbox", default="proof")
    setup.set_defaults(func=cmd_setup)

    checks = sub.add_parser("checks")
    checks.add_argument("--domain", required=True)
    checks.add_argument("--ip", default="")
    checks.add_argument("--tries", type=int, default=10)
    checks.add_argument("--wait", type=int, default=30)
    checks.set_defaults(func=cmd_checks)

    trip = sub.add_parser("roundtrip")
    trip.add_argument("--host", required=True)
    trip.add_argument("--user", required=True)
    trip.add_argument("--password", default="")
    trip.add_argument("--smtp-port", type=int, default=587)
    trip.add_argument("--imap-port", type=int, default=993)
    trip.add_argument("--timeout", type=int, default=180)
    trip.set_defaults(func=cmd_roundtrip)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
