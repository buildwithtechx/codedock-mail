import { Gauge } from 'lucide-react';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { useMailboxUsage } from '#/hooks/use-mail';
import type { Mailbox } from '#/interfaces/mail';
import { formatBytes } from '#/lib/utils';

export function UsageChecker({ domainId, mailbox }: { domainId: string; mailbox: Mailbox }) {
  const [open, setOpen] = useState(false);
  const [password, setPassword] = useState('');
  const usage = useMailboxUsage(domainId, mailbox.id);

  return (
    <span className="flex items-center gap-2">
      {usage.data && (
        <span className="text-muted-foreground text-xs">
          {formatBytes(usage.data.usedBytes)} used
          {mailbox.quotaMb > 0 && ` of ${mailbox.quotaMb} MB`}
        </span>
      )}
      {open ? (
        <span className="flex items-center gap-2">
          <Input
            className="h-8 w-32"
            type="password"
            autoComplete="off"
            placeholder="Password"
            aria-label={`Password for ${mailbox.address}`}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <Button
            variant="outline"
            size="sm"
            disabled={password === '' || usage.isPending}
            onClick={() =>
              usage.mutate(password, {
                onSuccess: () => {
                  setPassword('');
                  setOpen(false);
                },
              })
            }
          >
            Check
          </Button>
        </span>
      ) : (
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={`Check usage for ${mailbox.address}`}
          onClick={() => setOpen(true)}
        >
          <Gauge className="h-4 w-4" />
        </Button>
      )}
    </span>
  );
}
