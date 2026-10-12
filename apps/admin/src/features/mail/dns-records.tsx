import { Badge } from '#/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card';
import { useMailDns } from '#/hooks/use-mail';

export function DnsRecords({ domainId }: { domainId: string }) {
  const { data: dns } = useMailDns(domainId);

  return (
    <Card className="border-border/60 bg-card">
      <CardHeader>
        <CardTitle>DNS records to publish</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        {(dns?.records ?? []).map((record, index) => (
          <div key={`${record.type}-${record.name}-${index}`} className="text-sm">
            <div className="flex items-center gap-2">
              <Badge variant="outline">{record.type}</Badge>
              <code className="font-mono">{record.name}</code>
              <span className="text-muted-foreground text-xs">{record.source}</span>
            </div>
            <code className="mt-1 block break-all rounded bg-muted p-2 font-mono text-xs">
              {record.value}
            </code>
          </div>
        ))}
        {(dns?.records ?? []).length === 0 && (
          <p className="text-muted-foreground text-sm">No records returned.</p>
        )}
        {dns?.engineError && (
          <p className="text-muted-foreground text-xs">Engine warning: {dns.engineError}</p>
        )}
      </CardContent>
    </Card>
  );
}
