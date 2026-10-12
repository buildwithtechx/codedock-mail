import { useCallback, useEffect, useState } from 'react';
import { ApiUnauthorized, clearToken, getToken, mailApi } from './api';
import { AuditPanel } from './audit-panel';
import { BackupsPanel } from './backups-panel';
import { DomainPanel } from './domain-panel';
import { LoginForm } from './login-form';

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
    return <LoginForm onDone={() => probe()} />;
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
      <BackupsPanel />
      <AuditPanel />
    </main>
  );
}
