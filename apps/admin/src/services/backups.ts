import type { BackupManifest } from '#/interfaces/mail';
import { apiClient } from '#/lib/api-client';
import { handleApiError } from '#/lib/error';

export const backupService = {
  listBackups: async (): Promise<BackupManifest[]> => {
    try {
      return await apiClient.get<BackupManifest[]>('/api/backups');
    } catch (err) {
      throw handleApiError(err);
    }
  },

  createBackup: async (): Promise<BackupManifest> => {
    try {
      return await apiClient.post<BackupManifest>('/api/backups');
    } catch (err) {
      throw handleApiError(err);
    }
  },

  restoreBackup: async (id: string): Promise<{ status: string }> => {
    try {
      return await apiClient.post<{ status: string }>(`/api/backups/${id}/restore`);
    } catch (err) {
      throw handleApiError(err);
    }
  },

  deleteBackup: async (id: string): Promise<void> => {
    try {
      await apiClient.delete<void>(`/api/backups/${id}`);
    } catch (err) {
      throw handleApiError(err);
    }
  },
};
