import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { useUpdateQuota } from '#/hooks/use-mail';
import type { Mailbox } from '#/interfaces/mail';

export function QuotaEditor({ domainId, mailbox }: { domainId: string; mailbox: Mailbox }) {
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
