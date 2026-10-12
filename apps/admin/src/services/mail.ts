import type {
  DkimKey,
  Mailbox,
  MailboxCredentials,
  MailboxUsage,
  MailCert,
  MailCheckReport,
  MailDnsResponse,
  MailDomain,
  SieveState,
} from '#/interfaces/mail';
import { apiClient } from '#/lib/api-client';
import { handleApiError } from '#/lib/error';

export const mailService = {
  listDomains: async (): Promise<MailDomain[]> => {
    try {
      return await apiClient.get<MailDomain[]>('/api/domains');
    } catch (err) {
      throw handleApiError(err);
    }
  },

  createDomain: async (name: string): Promise<MailDomain> => {
    try {
      return await apiClient.post<MailDomain>('/api/domains', { name });
    } catch (err) {
      throw handleApiError(err);
    }
  },

  deleteDomain: async (id: string): Promise<void> => {
    try {
      await apiClient.delete<void>(`/api/domains/${id}`);
    } catch (err) {
      throw handleApiError(err);
    }
  },

  listMailboxes: async (domainId: string): Promise<Mailbox[]> => {
    try {
      return await apiClient.get<Mailbox[]>(`/api/domains/${domainId}/mailboxes`);
    } catch (err) {
      throw handleApiError(err);
    }
  },

  createMailbox: async (
    domainId: string,
    localPart: string,
    quotaMb: number
  ): Promise<MailboxCredentials> => {
    try {
      return await apiClient.post<MailboxCredentials>(`/api/domains/${domainId}/mailboxes`, {
        localPart,
        quotaMb,
      });
    } catch (err) {
      throw handleApiError(err);
    }
  },

  deleteMailbox: async (domainId: string, mailboxId: string): Promise<void> => {
    try {
      await apiClient.delete<void>(`/api/domains/${domainId}/mailboxes/${mailboxId}`);
    } catch (err) {
      throw handleApiError(err);
    }
  },

  updateQuota: async (domainId: string, mailboxId: string, quotaMb: number): Promise<Mailbox> => {
    try {
      return await apiClient.put<Mailbox>(`/api/domains/${domainId}/mailboxes/${mailboxId}/quota`, {
        quotaMb,
      });
    } catch (err) {
      throw handleApiError(err);
    }
  },

  createDkim: async (domainId: string): Promise<DkimKey> => {
    try {
      return await apiClient.post<DkimKey>(`/api/domains/${domainId}/dkim`, {
        selector: 'mail',
        algorithm: 'Ed25519',
      });
    } catch (err) {
      throw handleApiError(err);
    }
  },

  getDns: async (domainId: string): Promise<MailDnsResponse> => {
    try {
      return await apiClient.get<MailDnsResponse>(`/api/domains/${domainId}/dns`);
    } catch (err) {
      throw handleApiError(err);
    }
  },

  getChecks: async (domainId: string): Promise<MailCheckReport> => {
    try {
      return await apiClient.get<MailCheckReport>(`/api/domains/${domainId}/checks`);
    } catch (err) {
      throw handleApiError(err);
    }
  },

  listCerts: async (): Promise<MailCert[]> => {
    try {
      return await apiClient.get<MailCert[]>('/api/certs');
    } catch (err) {
      throw handleApiError(err);
    }
  },

  getUsage: async (
    domainId: string,
    mailboxId: string,
    password: string
  ): Promise<MailboxUsage> => {
    try {
      return await apiClient.get<MailboxUsage>(
        `/api/domains/${domainId}/mailboxes/${mailboxId}/usage`,
        {
          headers: { 'X-Mailbox-Password': password },
        }
      );
    } catch (err) {
      throw handleApiError(err);
    }
  },

  getSieve: async (domainId: string, mailboxId: string, password: string): Promise<SieveState> => {
    try {
      return await apiClient.get<SieveState>(
        `/api/domains/${domainId}/mailboxes/${mailboxId}/sieve`,
        {
          headers: { 'X-Mailbox-Password': password },
        }
      );
    } catch (err) {
      throw handleApiError(err);
    }
  },

  putSieve: async (
    domainId: string,
    mailboxId: string,
    payload: { password: string; name: string; script: string }
  ): Promise<{ status: string }> => {
    try {
      return await apiClient.put<{ status: string }>(
        `/api/domains/${domainId}/mailboxes/${mailboxId}/sieve`,
        payload
      );
    } catch (err) {
      throw handleApiError(err);
    }
  },
};
