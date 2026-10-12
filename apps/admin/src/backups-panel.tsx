import { useState } from 'react';
import { type BackupManifest, mailApi } from './api';
import { useAsync } from './use-async';

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function BackupsPanel() {
  const [notice, setNotice] = useState<string | null>(null);
  const backups = useAsync(() => mailApi.listBackups(), []);
  const fail = (err: unknown) => setNotice(err instanceof Error ? err.message : 'failed');

  return (
    <section>
      <h2>Backups</h2>
      {notice && <p className="error">{notice}</p>}
      {backups.error && <p className="error">{backups.error}</p>}
      <div className="row">
        <button
          type="button"
          onClick={() =>
            mailApi
              .createBackup()
              .then(() => {
                setNotice(null);
                backups.reload();
              })
              .catch(fail)
          }
        >
          Take backup
        </button>
      </div>
      <ul>
        {(backups.data ?? []).map((backup: BackupManifest) => (
          <li key={backup.id} className="row">
            <span>
              <code>{backup.id}</code>{' '}
              <span className="muted">
                {new Date(backup.createdAt).toLocaleString()} · db {formatBytes(backup.sizeBytes)}
                {backup.dataPath
                  ? ` · mail ${formatBytes(backup.dataSizeBytes ?? 0)}`
                  : ' · db only'}
              </span>
            </span>
            <span className="row">
              <button
                type="button"
                onClick={() => {
                  if (!window.confirm(`Restore database from backup ${backup.id}?`)) return;
                  mailApi
                    .restoreBackup(backup.id)
                    .then(() => setNotice(`Database restored from ${backup.id}`))
                    .catch(fail);
                }}
              >
                Restore db
              </button>
              <button
                type="button"
                onClick={() =>
                  mailApi
                    .deleteBackup(backup.id)
                    .then(() => backups.reload())
                    .catch(fail)
                }
              >
                Delete
              </button>
            </span>
          </li>
        ))}
      </ul>
      <p className="muted">
        Mail archives restore by hand: stop the engine, extract the mail archive over its data
        directory, start the engine.
      </p>
    </section>
  );
}
