import type { QueryClient } from '@tanstack/react-query';
import { createRootRouteWithContext, HeadContent, Outlet } from '@tanstack/react-router';
import { ThemeProvider } from '#/components/theme-provider';
import { Toaster } from '#/components/ui/sonner';

interface MyRouterContext {
  queryClient: QueryClient;
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  head: () => ({
    meta: [
      {
        name: 'description',
        content: 'codedock-mail admin - domains, mailboxes, DNS and deliverability',
      },
      { title: 'codedock-mail admin' },
    ],
  }),
  component: RootDocument,
});

function RootDocument() {
  return (
    <ThemeProvider>
      <HeadContent />
      <Outlet />
      <Toaster />
    </ThemeProvider>
  );
}
