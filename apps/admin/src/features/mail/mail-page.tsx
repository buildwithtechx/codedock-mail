import { Link } from '@tanstack/react-router';
import { ChevronRight, Globe, Plus, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { PageHeader } from '#/components/layout/page-header';
import { Button } from '#/components/ui/button';
import { Card, CardContent } from '#/components/ui/card';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { QueryErrorState } from '#/components/ui/query-error-state';
import { useCreateMailDomain, useDeleteMailDomain, useListMailDomains } from '#/hooks/use-mail';

export function MailPage() {
  const [newDomain, setNewDomain] = useState('');
  const { data: domains = [], isLoading, isError, refetch } = useListMailDomains();
  const createDomain = useCreateMailDomain();
  const deleteDomain = useDeleteMailDomain();

  if (isError) {
    return (
      <div className="space-y-6 pb-12">
        <PageHeader
          title="Mail"
          description="Self-hosted mail domains, mailboxes, DNS and deliverability."
        />
        <QueryErrorState
          title="Mail unavailable"
          description="Mail domains could not be loaded."
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6 pb-12">
      <PageHeader
        title="Mail"
        description="Self-hosted mail domains, mailboxes, DNS and deliverability."
      />
      <Card className="border-border/60 bg-card">
        <CardContent className="flex gap-2 p-4">
          <div className="flex-1 space-y-2">
            <Label htmlFor="mail-new-domain">New domain</Label>
            <Input
              id="mail-new-domain"
              placeholder="example.com"
              value={newDomain}
              onChange={(e) => setNewDomain(e.target.value)}
            />
          </div>
          <Button
            className="mt-7 gap-2"
            disabled={newDomain.trim() === '' || createDomain.isPending}
            onClick={() => {
              createDomain.mutate(newDomain.trim(), { onSuccess: () => setNewDomain('') });
            }}
          >
            <Plus className="h-4 w-4" />
            Add domain
          </Button>
        </CardContent>
      </Card>
      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading…</p>
      ) : domains.length === 0 ? (
        <Card className="border-border/60 bg-card">
          <CardContent className="p-6 text-muted-foreground text-sm">
            No domains yet. Add one to manage its mailboxes, DNS and checks.
          </CardContent>
        </Card>
      ) : (
        <Card className="border-border/60 bg-card">
          <CardContent className="space-y-2 p-4">
            {domains.map((domain) => (
              <div
                key={domain.id}
                className="flex w-full items-center justify-between rounded-lg border border-border/60 p-3"
              >
                <Link
                  to="/domains/$domainId"
                  params={{ domainId: domain.id }}
                  className="flex flex-1 items-center gap-2"
                >
                  <Globe className="h-4 w-4 text-muted-foreground" />
                  <span className="font-mono text-sm">{domain.name}</span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </Link>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Remove ${domain.name}`}
                  disabled={deleteDomain.isPending}
                  onClick={() => deleteDomain.mutate(domain.id)}
                >
                  <Trash2 className="h-4 w-4 text-muted-foreground" />
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
