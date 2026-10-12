import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';
import { AppLayout } from '#/components/layout/app-layout';
import { useAuthStore } from '#/stores/auth-store';

export const Route = createFileRoute('/_app')({
  beforeLoad: () => {
    if (!useAuthStore.getState().unlocked) {
      throw redirect({ to: '/login' });
    }
  },
  component: AppLayoutRoute,
});

function AppLayoutRoute() {
  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  );
}
