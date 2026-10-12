import { type MailCheck, mailApi } from './api';
import { useAsync } from './use-async';

export function ChecksPanel({ domainId }: { domainId: string }) {
  const checks = useAsync(() => mailApi.getChecks(domainId), [domainId]);

  return (
    <section>
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
    </section>
  );
}
