import { type AuditEntry, mailApi } from './api';
import { useAsync } from './use-async';

export function AuditPanel() {
  const audit = useAsync(() => mailApi.listAudit(50), []);

  return (
    <section>
      <h2>
        Audit log{' '}
        <button type="button" onClick={() => audit.reload()}>
          Refresh
        </button>
      </h2>
      {audit.error && <p className="error">{audit.error}</p>}
      <ul>
        {(audit.data ?? []).map((entry: AuditEntry) => (
          <li key={entry.id}>
            <span className="muted">{new Date(entry.createdAt).toLocaleString()}</span>{' '}
            <code>{entry.action}</code> <span className="muted">{entry.resourceType}</span>{' '}
            <code>{entry.resourceId}</code>{' '}
            {entry.detail && <span className="muted">{entry.detail}</span>}
          </li>
        ))}
      </ul>
    </section>
  );
}
