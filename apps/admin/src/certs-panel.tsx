import { type MailCert, mailApi } from './api';
import { useAsync } from './use-async';

export function CertsPanel() {
  const certs = useAsync(() => mailApi.listCerts(), []);

  return (
    <section>
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
