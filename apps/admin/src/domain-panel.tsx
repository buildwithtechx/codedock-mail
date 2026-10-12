import { useState } from 'react';
import { type Mailbox, mailApi } from './api';
import { CertsPanel } from './certs-panel';
import { ChecksPanel } from './checks-panel';
import { DnsPanel } from './dns-panel';
import { MailboxList } from './mailbox-list';
import { SievePanel } from './sieve-panel';

export function DomainPanel({ domainId, domainName }: { domainId: string; domainName: string }) {
  const [localPart, setLocalPart] = useState('');
  const [quotaMb, setQuotaMb] = useState('0');
  const [password, setPassword] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [dnsNonce, setDnsNonce] = useState(0);
  const [mailboxNonce, setMailboxNonce] = useState(0);
  const [sieveMailbox, setSieveMailbox] = useState<Mailbox | null>(null);
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
      <div className="row">
        <input
          placeholder="alice"
          value={localPart}
          onChange={(e) => setLocalPart(e.target.value)}
        />
        <input
          className="quota"
          inputMode="numeric"
          aria-label="Quota in MB, 0 for unlimited"
          placeholder="MB"
          value={quotaMb}
          onChange={(e) => setQuotaMb(e.target.value)}
        />
        <button
          type="button"
          onClick={() => {
            const parsed = Number(quotaMb);
            if (!Number.isInteger(parsed) || parsed < 0) {
              fail(new Error('quota must be a whole number of MB, 0 for unlimited'));
              return;
            }
            mailApi
              .createMailbox(domainId, localPart.trim(), parsed)
              .then((created) => {
                setLocalPart('');
                setPassword(created.password ?? null);
                setNotice(null);
                setMailboxNonce((n) => n + 1);
              })
              .catch(fail);
          }}
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
                setDnsNonce((n) => n + 1);
              })
              .catch(fail)
          }
        >
          DKIM
        </button>
      </div>
      <MailboxList
        key={mailboxNonce}
        domainId={domainId}
        onSelectSieve={(mailbox) => setSieveMailbox(mailbox)}
      />
      {sieveMailbox && (
        <SievePanel key={sieveMailbox.id} domainId={domainId} mailbox={sieveMailbox} />
      )}
      <DnsPanel domainId={domainId} nonce={dnsNonce} />
      <ChecksPanel domainId={domainId} />
      <CertsPanel />
    </section>
  );
}
