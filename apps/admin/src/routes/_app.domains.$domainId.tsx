import { createFileRoute, Link } from '@tanstack/react-router';
import { ChevronLeft } from 'lucide-react';
import { PageHeader } from '#/components/layout/page-header';
import { QueryErrorState } from '#/components/ui/query-error-state';
import { MailDomainPanel } from '#/features/mail/mail-domain-panel';
import { useListMailDomains } from '#/hooks/use-mail';

export const Route = createFileRoute('/_app/domains/$domainId')({
  component: DomainDetailPage,
});

function DomainDetailPage() {
  const { domainId } = Route.useParams();
  const { data: domains = [], isLoading, isError, refetch } = useListMailDomains();
  const domain = domains.find((entry) => entry.id === domainId) ?? null;

  if (isError) {
    return (
      <div className="space-y-6 pb-12">
        <PageHeader title="Domain" description="Mailboxes, DNS and deliverability." />
        <QueryErrorState
          title="Domain unavailable"
          description="The domain could not be loaded."
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  if (isLoading) {
    return <p className="text-muted-foreground text-sm">Loading…</p>;
  }

  if (!domain) {
    return (
      <div className="space-y-6 pb-12">
        <PageHeader title="Domain not found" description="It may have been deleted." />
        <Link
          to="/"
          className="inline-flex items-center gap-1 text-primary text-sm hover:underline"
        >
          <ChevronLeft className="h-4 w-4" />
          Back to domains
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6 pb-12">
      <PageHeader
        title={domain.name}
        description="Mailboxes, DNS and deliverability."
        action={
          <Link
            to="/"
            className="inline-flex items-center gap-1 text-muted-foreground text-sm hover:text-foreground"
          >
            <ChevronLeft className="h-4 w-4" />
            All domains
          </Link>
        }
      />
      <MailDomainPanel domainId={domain.id} domainName={domain.name} />
    </div>
  );
}
