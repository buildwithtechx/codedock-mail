import { DatabaseBackup, RotateCcw, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { PageHeader } from '#/components/layout/page-header';
import { Button } from '#/components/ui/button';
import { Card, CardContent } from '#/components/ui/card';
import { QueryErrorState } from '#/components/ui/query-error-state';
import {
  useCreateBackup,
  useDeleteBackup,
  useListBackups,
  useRestoreBackup,
} from '#/hooks/use-backups';
import { formatBytes } from '#/lib/utils';

export function BackupsPage() {
  const { data: backups = [], isLoading, isError, refetch } = useListBackups();
  const createBackup = useCreateBackup();
  const restoreBackup = useRestoreBackup();
  const deleteBackup = useDeleteBackup();

  if (isError) {
    return (
      <div className="space-y-6 pb-12">
        <PageHeader title="Backups" description="Database dumps plus mail-data archives." />
        <QueryErrorState
          title="Backups unavailable"
          description="Backups could not be loaded."
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6 pb-12">
      <PageHeader
        title="Backups"
        description="Database dumps plus mail-data archives."
        action={
          <Button
            className="gap-2"
            disabled={createBackup.isPending}
            onClick={() => createBackup.mutate()}
          >
            <DatabaseBackup className="h-4 w-4" />
            Take backup
          </Button>
        }
      />
      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading…</p>
      ) : backups.length === 0 ? (
        <Card className="border-border/60 bg-card">
          <CardContent className="p-6 text-muted-foreground text-sm">
            No backups yet. Take one before risky changes.
          </CardContent>
        </Card>
      ) : (
        <Card className="border-border/60 bg-card">
          <CardContent className="space-y-2 p-4">
            {backups.map((backup) => (
              <div
                key={backup.id}
                className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-border/60 p-3"
              >
                <span>
                  <code className="font-mono text-sm">{backup.id}</code>{' '}
                  <span className="text-muted-foreground text-xs">
                    {new Date(backup.createdAt).toLocaleString()} · db{' '}
                    {formatBytes(backup.sizeBytes)}
                    {backup.dataPath
                      ? ` · mail ${formatBytes(backup.dataSizeBytes ?? 0)}`
                      : ' · db only'}
                  </span>
                </span>
                <span className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    className="gap-2"
                    disabled={restoreBackup.isPending}
                    onClick={() => {
                      if (!window.confirm(`Restore database from backup ${backup.id}?`)) return;
                      restoreBackup.mutate(backup.id, {
                        onSuccess: () => toast.success(`Database restored from ${backup.id}`),
                      });
                    }}
                  >
                    <RotateCcw className="h-4 w-4" />
                    Restore db
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`Delete backup ${backup.id}`}
                    disabled={deleteBackup.isPending}
                    onClick={() => deleteBackup.mutate(backup.id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </span>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
      <p className="text-muted-foreground text-xs">
        Mail archives restore by hand: stop the engine, extract the mail archive over its data
        directory, start the engine.
      </p>
    </div>
  );
}
