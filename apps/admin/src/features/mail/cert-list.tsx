import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card';
import { useMailCerts } from '#/hooks/use-mail';

export function CertList() {
  const { data: certs = [] } = useMailCerts();

  return (
    <Card className="border-border/60 bg-card">
      <CardHeader>
        <CardTitle>Tracked certificates</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        <ul className="space-y-2">
          {certs.map((cert) => (
            <li key={cert.id} className="text-sm">
              <code className="font-mono">{cert.hostname}</code>{' '}
              <span className="text-muted-foreground text-xs">expires {cert.expiresAt}</span>
            </li>
          ))}
        </ul>
        {certs.length === 0 && (
          <p className="text-muted-foreground text-sm">No certificates tracked.</p>
        )}
      </CardContent>
    </Card>
  );
}
