import { type DependencyList, useCallback, useEffect, useRef, useState } from 'react';
import {
  ApiUnauthorized,
  clearToken,
  getToken,
  type Mailbox,
  type MailCert,
  type MailCheck,
  type MailDnsRecord,
  mailApi,
  setToken,
} from './api';

function useAsync<T>(
  run: () => Promise<T>,
  deps: DependencyList
): {
  data: T | null;
  error: string | null;
  reload: () => void;
} {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [nonce, setNonce] = useState(0);
  const runRef = useRef(run);
  runRef.current = run;
  // biome-ignore lint/correctness/useExhaustiveDependencies: nonce is a manual reload signal, not a reactive value
  useEffect(() => {
    let live = true;
    runRef
      .current()
      .then((result) => {
        if (live) {
          setData(result);
          setError(null);
        }
      })
      .catch((err: unknown) => {
        if (live) setError(err instanceof Error ? err.message : 'failed');
      });
    return () => {
      live = false;
    };
  }, [...deps, nonce]);
  return { data, error, reload: () => setNonce((n) => n + 1) };
}

function Login({ onDone }: { onDone: () => void }) {
  const [token, setTokenValue] = useState('');
  return (
    <main className="wrap">
      <h1>codedock-mail</h1>
      <p className="muted">
        Enter the maild API token (MAILD_API_TOKEN), or continue without one if the server runs
        open.
      </p>
      <div className="row">
        <input
          type="password"
          placeholder="API token"
          value={token}
          onChange={(e) => setTokenValue(e.target.value)}
        />
        <button
          type="button"
          onClick={() => {
            if (token.trim() !== '') setToken(token.trim());
            onDone();
          }}
        >
          Continue
        </button>
      </div>
    </main>
  );
}

function DomainPanel({ domainId, domainName }: { domainId: string; domainName: string }) {
  const [localPart, setLocalPart] = useState('');
  const [password, setPassword] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const mailboxes = useAsync(() => mailApi.listMailboxes(domainId), [domainId]);
  const dns = useAsync(() => mailApi.getDns(domainId), [domainId]);
  const checks = useAsync(() => mailApi.getChecks(domainId), [domainId]);
  const certs = useAsync(() => mailApi.listCerts(), [domainId]);

  const fail = (err: unknown) => setNotice(err instanceof Error ? err.message : 'failed');

  return (
    <section>
      <h2>{domainName}</h2>
      {notice && <p className="error">{notice}</p>}
      {password && (
        <p className="once">
          Mailbox created. Password shown once, never stored: <code>{password}</code>
        </p>
      )}
      <h3>Mailboxes</h3>
      <div className="row">
        <input
          placeholder="alice"
          value={localPart}
          onChange={(e) => setLocalPart(e.target.value)}
        />
        <button
          type="button"
          onClick={() =>
            mailApi
              .createMailbox(domainId, localPart.trim())
              .then((created: { password?: string }) => {
                setLocalPart('');
                setPassword(created.password ?? null);
                setNotice(null);
                mailboxes.reload();
              })
              .catch(fail)
          }
        >
          Add
        </button>
        <button
          type="button"
          onClick={() =>
            mailApi
              .createDkim(domainId)
              .then(() => {
                setNotice('DKIM key created');
                dns.reload();
              })
              .catch(fail)
          }
        >
          DKIM
        </button>
      </div>
      {mailboxes.error && <p className="error">{mailboxes.error}</p>}
      <ul>
        {(mailboxes.data ?? []).map((mailbox: Mailbox) => (
          <li key={mailbox.id} className="row">
            <code>{mailbox.address}</code>
            <button
              type="button"
              onClick={() =>
                mailApi
                  .deleteMailbox(domainId, mailbox.id)
                  .then(() => mailboxes.reload())
                  .catch(fail)
              }
            >
              Delete
            </button>
          </li>
        ))}
      </ul>
      <h3>DNS records to publish</h3>
      {dns.error && <p className="error">{dns.error}</p>}
      <ul>
        {(dns.data?.records ?? []).map((record: MailDnsRecord) => (
          <li key={`${record.type}-${record.name}-${record.value}`}>
            <code>
              {record.type} {record.name} {record.value}
            </code>{' '}
            <span className="muted">({record.source})</span>
          </li>
        ))}
      </ul>
      <h3>
        Checks{' '}
        <button type="button" onClick={() => checks.reload()}>
          Re-run
        </button>
      </h3>
      {checks.error && <p className="error">{checks.error}</p>}
      <ul>
        {(checks.data?.checks ?? []).map((check: MailCheck) => (
          <li key={check.name}>
            {check.passed ? '✓' : '✗'} <strong>{check.name}</strong>{' '}
            <span className="muted">{check.detail}</span>
          </li>
        ))}
      </ul>
      <h3>Tracked certificates</h3>
      {certs.error && <p className="error">{certs.error}</p>}
      <ul>
        {(certs.data ?? []).map((cert: MailCert) => (
          <li key={cert.id}>
            <code>{cert.hostname}</code> <span className="muted">expires {cert.expiresAt}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}

export function App() {
  const [authed, setAuthed] = useState<'unknown' | 'yes' | 'no'>('unknown');
  const [domains, setDomains] = useState<{ id: string; name: string }[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [newDomain, setNewDomain] = useState('');
  const [error, setError] = useState<string | null>(null);

  const probe = useCallback(() => {
    mailApi
      .listDomains()
      .then((result) => {
        setDomains(result);
        setError(null);
        setAuthed('yes');
        setActiveId((current) => current ?? result[0]?.id ?? null);
      })
      .catch((err: unknown) => {
        if (err instanceof ApiUnauthorized || (!getToken() && err instanceof Error))
          setAuthed('no');
        else setError(err instanceof Error ? err.message : 'failed');
      });
  }, []);

  useEffect(() => {
    probe();
  }, [probe]);

  if (authed === 'unknown') {
    return (
      <main className="wrap">
        <h1>codedock-mail</h1>
        <p className="muted">Connecting…</p>
      </main>
    );
  }

  if (authed === 'no') {
    return <Login onDone={() => probe()} />;
  }

  const active = domains.find((domain) => domain.id === activeId) ?? domains[0] ?? null;

  return (
    <main className="wrap">
      <header className="row">
        <h1>codedock-mail</h1>
        <button
          type="button"
          onClick={() => {
            clearToken();
            setAuthed('no');
          }}
        >
          Sign out
        </button>
      </header>
      {error && <p className="error">{error}</p>}
      <div className="row">
        <input
          placeholder="example.com"
          value={newDomain}
          onChange={(e) => setNewDomain(e.target.value)}
        />
        <button
          type="button"
          onClick={() =>
            mailApi
              .createDomain(newDomain.trim())
              .then(() => {
                setNewDomain('');
                probe();
              })
              .catch((err: unknown) => setError(err instanceof Error ? err.message : 'failed'))
          }
        >
          Add domain
        </button>
      </div>
      <div className="row">
        {domains.map((domain) => (
          <button
            type="button"
            key={domain.id}
            className={active?.id === domain.id ? 'active' : ''}
            onClick={() => setActiveId(domain.id)}
          >
            {domain.name}
          </button>
        ))}
      </div>
      {active && <DomainPanel key={active.id} domainId={active.id} domainName={active.name} />}
    </main>
  );
}
