import { Plus, ShieldAlert, SlidersHorizontal, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { QueryErrorState } from '#/components/ui/query-error-state';
import {
  useCreateDkim,
  useCreateMailbox,
  useDeleteMailbox,
  useListMailboxes,
  useUpdateQuota,
} from '#/hooks/use-mail';
import type { Mailbox } from '#/interfaces/mail';

function QuotaEditor({ domainId, mailbox }: { domainId: string; mailbox: Mailbox }) {
  const [quota, setQuota] = useState(String(mailbox.quotaMb));
  const updateQuota = useUpdateQuota(domainId);

  return (
    <span className="flex items-center gap-2">
      <Input
        className="h-8 w-20"
        inputMode="numeric"
        aria-label={`Quota in MB for ${mailbox.address}`}
        value={quota}
        onChange={(e) => setQuota(e.target.value)}
      />
      <Button
        variant="outline"
        size="sm"
        disabled={updateQuota.isPending}
        onClick={() => {
          const parsed = Number(quota);
          if (!Number.isInteger(parsed) || parsed < 0) return;
          updateQuota.mutate({ mailboxId: mailbox.id, quotaMb: parsed });
        }}
      >
        Save
      </Button>
    </span>
  );
}

export function MailboxManager({
  domainId,
  domainName,
  onPassword,
  onSelectSieve,
}: {
  domainId: string;
  domainName: string;
  onPassword: (password: string | null) => void;
  onSelectSieve: (mailboxId: string) => void;
}) {
  const [localPart, setLocalPart] = useState('');
  const [quotaMb, setQuotaMb] = useState('0');
  const { data: mailboxes = [], isLoading, isError, refetch } = useListMailboxes(domainId);
  const createMailbox = useCreateMailbox(domainId);
  const deleteMailbox = useDeleteMailbox(domainId);
  const createDkim = useCreateDkim(domainId);

  if (isError) {
    return (
      <QueryErrorState
        title="Mailboxes unavailable"
        description="The mail server could not be reached."
        onRetry={() => refetch()}
      />
    );
  }

  return (
    <Card className="border-border/60 bg-card">
      <CardHeader>
        <CardTitle>Mailboxes on {domainName}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex flex-wrap gap-2">
          <div className="min-w-32 flex-1 space-y-2">
            <Label htmlFor="mailbox-local">New mailbox</Label>
            <Input
              id="mailbox-local"
              placeholder="alice"
              value={localPart}
              onChange={(e) => setLocalPart(e.target.value)}
            />
          </div>
          <div className="w-28 space-y-2">
            <Label htmlFor="mailbox-quota">MB (0 = ∞)</Label>
            <Input
              id="mailbox-quota"
              inputMode="numeric"
              value={quotaMb}
              onChange={(e) => setQuotaMb(e.target.value)}
            />
          </div>
          <Button
            className="mt-7 gap-2"
            disabled={localPart.trim() === '' || createMailbox.isPending}
            onClick={() => {
              const parsed = Number(quotaMb);
              if (!Number.isInteger(parsed) || parsed < 0) return;
              createMailbox.mutate(
                { localPart: localPart.trim(), quotaMb: parsed },
                {
                  onSuccess: (credentials) => {
                    setLocalPart('');
                    onPassword(credentials.password ?? null);
                  },
                }
              );
            }}
          >
            <Plus className="h-4 w-4" />
            Add
          </Button>
          <Button
            className="mt-7 gap-2"
            variant="outline"
            disabled={createDkim.isPending}
            onClick={() => createDkim.mutate()}
          >
            <ShieldAlert className="h-4 w-4" />
            DKIM
          </Button>
        </div>
        {isLoading ? (
          <p className="text-muted-foreground text-sm">Loading…</p>
        ) : (
          <ul className="divide-y divide-border/60">
            {mailboxes.map((mailbox) => (
              <li
                key={mailbox.id}
                className="flex flex-wrap items-center justify-between gap-2 py-2"
              >
                <span className="font-mono text-sm">{mailbox.address}</span>
                <span className="flex items-center gap-2">
                  <QuotaEditor domainId={domainId} mailbox={mailbox} />
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`Filters for ${mailbox.address}`}
                    onClick={() => onSelectSieve(mailbox.id)}
                  >
                    <SlidersHorizontal className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`Delete ${mailbox.address}`}
                    disabled={deleteMailbox.isPending}
                    onClick={() => deleteMailbox.mutate(mailbox.id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </span>
              </li>
            ))}
            {mailboxes.length === 0 && (
              <li className="py-2 text-muted-foreground text-sm">No mailboxes yet.</li>
            )}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
