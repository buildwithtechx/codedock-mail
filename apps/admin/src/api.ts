export interface MailDomain {
  id: string;
  name: string;
}

export interface Mailbox {
  id: string;
  address: string;
  quotaMb: number;
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

export interface SieveScript {
  name: string;
  active: boolean;
}

export interface SieveState {
  address: string;
  scripts: SieveScript[];
  activeScript: string;
}

export interface BackupManifest {
  id: string;
  createdAt: string;
  path: string;
  sizeBytes: number;
  dataPath?: string;
  dataSizeBytes?: number;
  tables: string[];
}

export interface AuditEntry {
  id: string;
  actor: string;
  action: string;
  resourceType: string;
  resourceId: string;
  detail: string;
  createdAt: string;
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
  createMailbox: (
    domainId: string,
    localPart: string,
    quotaMb: number
  ): Promise<MailboxCredentials> =>
    request(`/api/domains/${domainId}/mailboxes`, {
      method: 'POST',
      body: JSON.stringify({ localPart, quotaMb }),
    }),
  deleteMailbox: (domainId: string, mailboxId: string): Promise<void> =>
    request(`/api/domains/${domainId}/mailboxes/${mailboxId}`, { method: 'DELETE' }),
  updateQuota: (domainId: string, mailboxId: string, quotaMb: number): Promise<Mailbox> =>
    request(`/api/domains/${domainId}/mailboxes/${mailboxId}/quota`, {
      method: 'PUT',
      body: JSON.stringify({ quotaMb }),
    }),
  getSieve: (domainId: string, mailboxId: string, password: string): Promise<SieveState> =>
    request(`/api/domains/${domainId}/mailboxes/${mailboxId}/sieve`, {
      headers: { 'X-Mailbox-Password': password },
    }),
  putSieve: (
    domainId: string,
    mailboxId: string,
    payload: { password: string; name: string; script: string }
  ): Promise<{ status: string }> =>
    request(`/api/domains/${domainId}/mailboxes/${mailboxId}/sieve`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),
  listBackups: (): Promise<BackupManifest[]> => request('/api/backups'),
  createBackup: (): Promise<BackupManifest> => request('/api/backups', { method: 'POST' }),
  restoreBackup: (id: string): Promise<{ status: string }> =>
    request(`/api/backups/${id}/restore`, { method: 'POST' }),
  deleteBackup: (id: string): Promise<void> => request(`/api/backups/${id}`, { method: 'DELETE' }),
  listAudit: (limit: number): Promise<AuditEntry[]> => request(`/api/audit?limit=${limit}`),
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
