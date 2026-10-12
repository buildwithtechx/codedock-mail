export interface MailDomain {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

export interface Mailbox {
  id: string;
  domainId: string;
  localPart: string;
  address: string;
  quotaMb: number;
  createdAt: string;
  updatedAt: string;
}

export interface MailboxCredentials extends Mailbox {
  password?: string;
}

export interface DkimKey {
  id: string;
  domainId: string;
  selector: string;
  algorithm: string;
  keyId?: string;
  publicKey?: string;
  dnsRecord?: string;
  createdAt: string;
}

export interface MailDnsRecord {
  type: string;
  name: string;
  value: string;
  source: string;
}

export interface MailDnsResponse {
  domain: string;
  records: MailDnsRecord[];
  engineError?: string;
}

export interface MailCheck {
  name: string;
  passed: boolean;
  detail: string;
}

export interface MailCheckReport {
  domain: string;
  checks: MailCheck[];
}

export interface MailCert {
  id: string;
  hostname: string;
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
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
