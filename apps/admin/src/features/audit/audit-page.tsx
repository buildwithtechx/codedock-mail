import { PageHeader } from '#/components/layout/page-header';
import { Button } from '#/components/ui/button';
import { Card, CardContent } from '#/components/ui/card';
import { QueryErrorState } from '#/components/ui/query-error-state';
import { useAuditLog } from '#/hooks/use-audit';

export function AuditPage() {
  const { data: entries = [], isLoading, isError, refetch } = useAuditLog(50);

  if (isError) {
    return (
      <div className="space-y-6 pb-12">
        <PageHeader title="Audit log" description="Newest first, operator actions." />
        <QueryErrorState
          title="Audit log unavailable"
          description="Entries could not be loaded."
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6 pb-12">
      <PageHeader
        title="Audit log"
        description="Newest first, operator actions."
        action={
          <Button variant="outline" onClick={() => refetch()}>
            Refresh
          </Button>
        }
      />
      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading…</p>
      ) : (
        <Card className="border-border/60 bg-card">
          <CardContent className="p-4">
            <ul className="space-y-2">
              {entries.map((entry) => (
                <li key={entry.id} className="text-sm">
                  <span className="text-muted-foreground text-xs">
                    {new Date(entry.createdAt).toLocaleString()}
                  </span>{' '}
                  <code className="font-mono">{entry.action}</code>{' '}
                  <span className="text-muted-foreground">{entry.resourceType}</span>{' '}
                  <code className="font-mono">{entry.resourceId}</code>{' '}
                  {entry.detail && (
                    <span className="text-muted-foreground text-xs">{entry.detail}</span>
                  )}
                </li>
              ))}
              {entries.length === 0 && (
                <li className="text-muted-foreground text-sm">No entries yet.</li>
              )}
            </ul>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
