import { useState } from 'react';
import { type Mailbox, mailApi } from './api';

export function SievePanel({ domainId, mailbox }: { domainId: string; mailbox: Mailbox }) {
  const [password, setPassword] = useState('');
  const [name, setName] = useState('main');
  const [script, setScript] = useState('');
  const [scripts, setScripts] = useState<string[]>([]);
  const [notice, setNotice] = useState<string | null>(null);
  const fail = (err: unknown) => setNotice(err instanceof Error ? err.message : 'failed');

  return (
    <section>
      <h3>Filters for {mailbox.address}</h3>
      {notice && <p className="error">{notice}</p>}
      {scripts.length > 0 && (
        <p className="muted">On server: {scripts.join(', ')}. Loading shows the active script.</p>
      )}
      <div className="row">
        <input
          type="password"
          placeholder="Mailbox password"
          autoComplete="off"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <input placeholder="Script name" value={name} onChange={(e) => setName(e.target.value)} />
        <button
          type="button"
          onClick={() =>
            mailApi
              .getSieve(domainId, mailbox.id, password)
              .then((state) => {
                setScripts(
                  state.scripts.map((entry) => entry.name + (entry.active ? ' (active)' : ''))
                );
                setScript(state.activeScript);
                setNotice(null);
              })
              .catch(fail)
          }
        >
          Load
        </button>
        <button
          type="button"
          onClick={() =>
            mailApi
              .putSieve(domainId, mailbox.id, { password, name: name.trim(), script })
              .then(() => setNotice('Filter saved and activated'))
              .catch(fail)
          }
        >
          Save and activate
        </button>
      </div>
      <textarea
        className="script"
        rows={10}
        placeholder={
          'require "fileinto";\n\nif header :contains "subject" "[newsletter]" {\n  fileinto "Newsletters";\n}'
        }
        value={script}
        onChange={(e) => setScript(e.target.value)}
      />
    </section>
  );
}
