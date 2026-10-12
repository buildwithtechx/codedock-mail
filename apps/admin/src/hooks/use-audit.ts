import { useQuery } from '@tanstack/react-query';
import { auditService } from '#/services/audit';

export const useAuditLog = (limit: number) =>
  useQuery({ queryKey: ['mail-audit', limit], queryFn: () => auditService.listAudit(limit) });
