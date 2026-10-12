import { Link, useNavigate, useRouterState } from '@tanstack/react-router';
import { Archive, LogOut, Mail, ScrollText } from 'lucide-react';
import type * as React from 'react';
import { Button } from '#/components/ui/button';
import { cn } from '#/lib/utils';
import { useAuthStore } from '#/stores/auth-store';

const NAV_ITEMS = [
  { to: '/', label: 'Domains', icon: Mail },
  { to: '/backups', label: 'Backups', icon: Archive },
  { to: '/audit', label: 'Audit', icon: ScrollText },
] as const;

export function AppLayout({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const clearToken = useAuthStore((state) => state.clearToken);

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="sticky top-0 z-40 border-border/60 border-b bg-background/80 backdrop-blur">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between gap-4 px-4 sm:px-6">
          <Link to="/" className="flex items-center gap-2 font-semibold tracking-tight">
            <Mail className="h-4 w-4 text-primary" />
            codedock-mail
          </Link>
          <nav className="flex items-center gap-1">
            {NAV_ITEMS.map((item) => (
              <Link
                key={item.to}
                to={item.to}
                className={cn(
                  'flex items-center gap-2 rounded-lg px-3 py-2 text-sm transition-colors',
                  pathname === item.to
                    ? 'bg-muted font-medium text-foreground'
                    : 'text-muted-foreground hover:bg-muted/70 hover:text-foreground'
                )}
              >
                <item.icon className="h-4 w-4" />
                {item.label}
              </Link>
            ))}
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label="Sign out"
              onClick={() => {
                clearToken();
                navigate({ to: '/login' });
              }}
            >
              <LogOut className="h-4 w-4" />
            </Button>
          </nav>
        </div>
      </header>
      <main className="page-transition mx-auto max-w-5xl space-y-6 px-4 py-6 pb-12 sm:px-6">
        {children}
      </main>
    </div>
  );
}
