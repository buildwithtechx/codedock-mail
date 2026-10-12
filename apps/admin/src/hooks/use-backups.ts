import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { backupService } from '#/services/backups';

export const useListBackups = () =>
  useQuery({ queryKey: ['mail-backups'], queryFn: () => backupService.listBackups() });

export const useCreateBackup = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => backupService.createBackup(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mail-backups'] }),
  });
};

export const useRestoreBackup = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => backupService.restoreBackup(id),
    onSuccess: () => qc.invalidateQueries(),
  });
};

export const useDeleteBackup = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => backupService.deleteBackup(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mail-backups'] }),
  });
};
