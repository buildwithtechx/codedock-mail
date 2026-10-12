import { type MailDnsRecord, mailApi } from './api';
import { useAsync } from './use-async';

export function DnsPanel({ domainId, nonce }: { domainId: string; nonce: number }) {
  const dns = useAsync(() => mailApi.getDns(domainId), [domainId, nonce]);

  return (
    <section>
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
    </section>
  );
}
