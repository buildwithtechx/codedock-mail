import { KeyRound } from 'lucide-react';
import { useState } from 'react';
import { Alert, AlertDescription } from '#/components/ui/alert';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import { CertList } from './cert-list';
import { DeliverabilityChecks } from './deliverability-checks';
import { DnsRecords } from './dns-records';
import { MailboxManager } from './mailbox-manager';
import { SieveEditor } from './sieve-editor';

export function MailDomainPanel({
  domainId,
  domainName,
}: {
  domainId: string;
  domainName: string;
}) {
  const [oneTimePassword, setOneTimePassword] = useState<string | null>(null);
  const [sieveMailboxId, setSieveMailboxId] = useState<string | null>(null);

  return (
    <div className="space-y-4">
      {oneTimePassword && (
        <Alert>
          <KeyRound className="h-4 w-4" />
          <AlertDescription>
            Mailbox created. The password is shown once and never stored:{' '}
            <code className="font-mono font-semibold">{oneTimePassword}</code>
          </AlertDescription>
        </Alert>
      )}
      <Tabs defaultValue="mailboxes">
        <TabsList>
          <TabsTrigger value="mailboxes">Mailboxes</TabsTrigger>
          <TabsTrigger value="dns">DNS</TabsTrigger>
          <TabsTrigger value="checks">Checks</TabsTrigger>
          <TabsTrigger value="certs">Certificates</TabsTrigger>
        </TabsList>
        <TabsContent value="mailboxes" className="space-y-4">
          <MailboxManager
            domainId={domainId}
            domainName={domainName}
            onPassword={(password) => setOneTimePassword(password)}
            onSelectSieve={(mailboxId) => setSieveMailboxId(mailboxId)}
          />
          {sieveMailboxId && (
            <SieveEditor
              key={sieveMailboxId}
              domainId={domainId}
              mailboxId={sieveMailboxId}
              onClose={() => setSieveMailboxId(null)}
            />
          )}
        </TabsContent>
        <TabsContent value="dns">
          <DnsRecords domainId={domainId} />
        </TabsContent>
        <TabsContent value="checks">
          <DeliverabilityChecks domainId={domainId} />
        </TabsContent>
        <TabsContent value="certs">
          <CertList />
        </TabsContent>
      </Tabs>
    </div>
  );
}
