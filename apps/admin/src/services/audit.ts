import type { AuditEntry } from '#/interfaces/mail';
import { apiClient } from '#/lib/api-client';
import { handleApiError } from '#/lib/error';

export const auditService = {
  listAudit: async (limit: number): Promise<AuditEntry[]> => {
    try {
      return await apiClient.get<AuditEntry[]>(`/api/audit?limit=${limit}`);
    } catch (err) {
      throw handleApiError(err);
    }
  },
};
