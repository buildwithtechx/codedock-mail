import { useState } from 'react';
import { type Mailbox, mailApi } from './api';
import { useAsync } from './use-async';

function QuotaEditor({
  domainId,
  mailbox,
  onSaved,
  onFailed,
}: {
  domainId: string;
  mailbox: Mailbox;
  onSaved: () => void;
  onFailed: (err: unknown) => void;
}) {
  const [quota, setQuota] = useState(String(mailbox.quotaMb));
  return (
    <span className="row">
      <input
        className="quota"
        inputMode="numeric"
        aria-label={`Quota in MB for ${mailbox.address}`}
        value={quota}
        onChange={(e) => setQuota(e.target.value)}
      />
      <button
        type="button"
        onClick={() => {
          const parsed = Number(quota);
          if (!Number.isInteger(parsed) || parsed < 0) {
            onFailed(new Error('quota must be a whole number of MB, 0 for unlimited'));
            return;
          }
          mailApi
            .updateQuota(domainId, mailbox.id, parsed)
            .then(() => onSaved())
            .catch(onFailed);
        }}
      >
        Save quota
      </button>
    </span>
  );
}

export function MailboxList({
  domainId,
  onSelectSieve,
}: {
  domainId: string;
  onSelectSieve: (mailbox: Mailbox) => void;
}) {
  const [notice, setNotice] = useState<string | null>(null);
  const mailboxes = useAsync(() => mailApi.listMailboxes(domainId), [domainId]);
  const fail = (err: unknown) => setNotice(err instanceof Error ? err.message : 'failed');

  return (
    <section>
      <h3>Mailboxes</h3>
      {notice && <p className="error">{notice}</p>}
      {mailboxes.error && <p className="error">{mailboxes.error}</p>}
      <ul>
        {(mailboxes.data ?? []).map((mailbox: Mailbox) => (
          <li key={mailbox.id} className="row">
            <code>{mailbox.address}</code>
            <QuotaEditor
              domainId={domainId}
              mailbox={mailbox}
              onSaved={() => {
                setNotice(null);
                mailboxes.reload();
              }}
              onFailed={fail}
            />
            <button type="button" onClick={() => onSelectSieve(mailbox)}>
              Filters
            </button>
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
    </section>
  );
}
