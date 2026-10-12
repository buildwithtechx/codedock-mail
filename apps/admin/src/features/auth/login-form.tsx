import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { useAuthStore } from '#/stores/auth-store';

export function LoginForm() {
  const navigate = useNavigate();
  const setToken = useAuthStore((state) => state.setToken);
  const [token, setTokenValue] = useState('');

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-sm border-border/60 bg-card">
        <CardHeader>
          <CardTitle>codedock-mail</CardTitle>
          <CardDescription>
            Enter the maild API token (MAILD_API_TOKEN), or continue without one if the server runs
            open.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="api-token">API token</Label>
            <Input
              id="api-token"
              type="password"
              placeholder="API token"
              value={token}
              onChange={(e) => setTokenValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  setToken(token.trim() === '' ? null : token.trim());
                  navigate({ to: '/' });
                }
              }}
            />
          </div>
          <Button
            className="w-full"
            onClick={() => {
              setToken(token.trim() === '' ? null : token.trim());
              navigate({ to: '/' });
            }}
          >
            Continue
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
