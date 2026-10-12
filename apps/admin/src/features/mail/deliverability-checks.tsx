import { CheckCircle2, XCircle } from 'lucide-react';
import { Button } from '#/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card';
import { useMailChecks } from '#/hooks/use-mail';

export function DeliverabilityChecks({ domainId }: { domainId: string }) {
  const { data: checks, refetch: refetchChecks } = useMailChecks(domainId);

  return (
    <Card className="border-border/60 bg-card">
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Deliverability checks</CardTitle>
        <Button variant="outline" size="sm" onClick={() => refetchChecks()}>
          Re-run
        </Button>
      </CardHeader>
      <CardContent className="space-y-2">
        {(checks?.checks ?? []).map((check) => (
          <div key={check.name} className="flex items-start gap-2 text-sm">
            {check.passed ? (
              <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />
            ) : (
              <XCircle className="mt-0.5 h-4 w-4 shrink-0 text-destructive" />
            )}
            <div>
              <p className="font-medium">{check.name}</p>
              <p className="break-all text-muted-foreground text-xs">{check.detail}</p>
            </div>
          </div>
        ))}
        {(checks?.checks ?? []).length === 0 && (
          <p className="text-muted-foreground text-sm">No checks returned.</p>
        )}
      </CardContent>
    </Card>
  );
}
