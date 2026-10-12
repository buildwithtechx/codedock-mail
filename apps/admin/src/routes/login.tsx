import { createFileRoute, redirect } from '@tanstack/react-router';
import { LoginForm } from '#/features/auth/login-form';
import { useAuthStore } from '#/stores/auth-store';

export const Route = createFileRoute('/login')({
  beforeLoad: () => {
    if (useAuthStore.getState().unlocked) {
      throw redirect({ to: '/' });
    }
  },
  component: LoginForm,
});
