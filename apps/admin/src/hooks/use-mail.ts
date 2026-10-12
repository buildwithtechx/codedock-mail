import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { mailService } from '#/services/mail';

export const useListMailDomains = () =>
  useQuery({ queryKey: ['mail-domains'], queryFn: () => mailService.listDomains() });

export const useCreateMailDomain = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => mailService.createDomain(name),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mail-domains'] }),
  });
};

export const useDeleteMailDomain = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (domainId: string) => mailService.deleteDomain(domainId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mail-domains'] }),
  });
};

export const useListMailboxes = (domainId?: string) =>
  useQuery({
    queryKey: ['mail-domains', domainId, 'mailboxes'],
    queryFn: () => (domainId ? mailService.listMailboxes(domainId) : Promise.resolve([])),
    enabled: Boolean(domainId),
  });

export const useCreateMailbox = (domainId?: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ localPart, quotaMb }: { localPart: string; quotaMb: number }) => {
      if (!domainId) throw new Error('No domain selected');
      return mailService.createMailbox(domainId, localPart, quotaMb);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mail-domains', domainId, 'mailboxes'] }),
  });
};

export const useDeleteMailbox = (domainId?: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (mailboxId: string) => {
      if (!domainId) throw new Error('No domain selected');
      return mailService.deleteMailbox(domainId, mailboxId);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mail-domains', domainId, 'mailboxes'] }),
  });
};

export const useUpdateQuota = (domainId?: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ mailboxId, quotaMb }: { mailboxId: string; quotaMb: number }) => {
      if (!domainId) throw new Error('No domain selected');
      return mailService.updateQuota(domainId, mailboxId, quotaMb);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mail-domains', domainId, 'mailboxes'] }),
  });
};

export const useCreateDkim = (domainId?: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => {
      if (!domainId) throw new Error('No domain selected');
      return mailService.createDkim(domainId);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mail-domains', domainId] }),
  });
};

export const useMailDns = (domainId?: string) =>
  useQuery({
    queryKey: ['mail-domains', domainId, 'dns'],
    queryFn: () => (domainId ? mailService.getDns(domainId) : Promise.resolve(null)),
    enabled: Boolean(domainId),
  });

export const useMailChecks = (domainId?: string) =>
  useQuery({
    queryKey: ['mail-domains', domainId, 'checks'],
    queryFn: () => (domainId ? mailService.getChecks(domainId) : Promise.resolve(null)),
    enabled: Boolean(domainId),
  });

export const useMailCerts = () =>
  useQuery({ queryKey: ['mail-certs'], queryFn: () => mailService.listCerts() });

export const useMailboxUsage = (domainId?: string, mailboxId?: string) =>
  useMutation({
    mutationFn: (password: string) => {
      if (!domainId || !mailboxId) throw new Error('No mailbox selected');
      return mailService.getUsage(domainId, mailboxId, password);
    },
  });

export const useLoadSieve = (domainId?: string, mailboxId?: string) =>
  useMutation({
    mutationFn: (password: string) => {
      if (!domainId || !mailboxId) throw new Error('No mailbox selected');
      return mailService.getSieve(domainId, mailboxId, password);
    },
  });

export const useSaveSieve = (domainId?: string, mailboxId?: string) =>
  useMutation({
    mutationFn: (payload: { password: string; name: string; script: string }) => {
      if (!domainId || !mailboxId) throw new Error('No mailbox selected');
      return mailService.putSieve(domainId, mailboxId, payload);
    },
  });
