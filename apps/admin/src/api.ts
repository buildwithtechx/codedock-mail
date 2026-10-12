export interface MailDomain {
  id: string;
  name: string;
}

export interface Mailbox {
  id: string;
  address: string;
}

export interface MailboxCredentials extends Mailbox {
  password?: string;
}

export interface MailDnsRecord {
  type: string;
  name: string;
  value: string;
  source: string;
}

export interface MailCheck {
  name: string;
  passed: boolean;
  detail: string;
}

export interface MailCert {
  id: string;
  hostname: string;
  expiresAt: string;
}

export class ApiUnauthorized extends Error {
  constructor() {
    super('unauthorized');
    this.name = 'ApiUnauthorized';
  }
}

const TOKEN_KEY = 'maild_token';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken();
  const response = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });
  if (response.status === 401) {
    clearToken();
    throw new ApiUnauthorized();
  }
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(body.error || 'request failed');
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export const mailApi = {
  listDomains: (): Promise<MailDomain[]> => request('/api/domains'),
  createDomain: (name: string): Promise<MailDomain> =>
    request('/api/domains', { method: 'POST', body: JSON.stringify({ name }) }),
  deleteDomain: (id: string): Promise<void> => request(`/api/domains/${id}`, { method: 'DELETE' }),
  listMailboxes: (domainId: string): Promise<Mailbox[]> =>
    request(`/api/domains/${domainId}/mailboxes`),
  createMailbox: (domainId: string, localPart: string): Promise<MailboxCredentials> =>
    request(`/api/domains/${domainId}/mailboxes`, {
      method: 'POST',
      body: JSON.stringify({ localPart }),
    }),
  deleteMailbox: (domainId: string, mailboxId: string): Promise<void> =>
    request(`/api/domains/${domainId}/mailboxes/${mailboxId}`, { method: 'DELETE' }),
  createDkim: (domainId: string): Promise<unknown> =>
    request(`/api/domains/${domainId}/dkim`, {
      method: 'POST',
      body: JSON.stringify({ selector: 'mail', algorithm: 'Ed25519' }),
    }),
  getDns: (domainId: string): Promise<{ records: MailDnsRecord[] }> =>
    request(`/api/domains/${domainId}/dns`),
  getChecks: (domainId: string): Promise<{ checks: MailCheck[] }> =>
    request(`/api/domains/${domainId}/checks`),
  listCerts: (): Promise<MailCert[]> => request('/api/certs'),
};
