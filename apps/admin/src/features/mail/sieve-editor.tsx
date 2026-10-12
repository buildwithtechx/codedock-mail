import { useState } from 'react';
import { toast } from 'sonner';
import { Button } from '#/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { Textarea } from '#/components/ui/textarea';
import { useLoadSieve, useSaveSieve } from '#/hooks/use-mail';

export function SieveEditor({
  domainId,
  mailboxId,
  onClose,
}: {
  domainId: string;
  mailboxId: string;
  onClose: () => void;
}) {
  const [password, setPassword] = useState('');
  const [name, setName] = useState('main');
  const [script, setScript] = useState('');
  const [scripts, setScripts] = useState<string[]>([]);
  const loadSieve = useLoadSieve(domainId, mailboxId);
  const saveSieve = useSaveSieve(domainId, mailboxId);

  return (
    <Card className="border-border/60 bg-card">
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Filters</CardTitle>
        <Button variant="ghost" size="sm" onClick={onClose}>
          Close
        </Button>
      </CardHeader>
      <CardContent className="space-y-3">
        {scripts.length > 0 && (
          <p className="text-muted-foreground text-xs">
            On server: {scripts.join(', ')}. Loading shows the active script.
          </p>
        )}
        <div className="flex flex-wrap gap-2">
          <div className="min-w-40 flex-1 space-y-2">
            <Label htmlFor="sieve-password">Mailbox password</Label>
            <Input
              id="sieve-password"
              type="password"
              autoComplete="off"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          <div className="min-w-32 flex-1 space-y-2">
            <Label htmlFor="sieve-name">Script name</Label>
            <Input id="sieve-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <Button
            className="mt-7"
            variant="outline"
            disabled={password === '' || loadSieve.isPending}
            onClick={() =>
              loadSieve.mutate(password, {
                onSuccess: (state) => {
                  setScripts(
                    state.scripts.map((entry) => entry.name + (entry.active ? ' (active)' : ''))
                  );
                  setScript(state.activeScript);
                },
              })
            }
          >
            Load
          </Button>
          <Button
            className="mt-7"
            disabled={password === '' || name.trim() === '' || saveSieve.isPending}
            onClick={() =>
              saveSieve.mutate(
                { password, name: name.trim(), script },
                { onSuccess: () => toast.success('Filter saved and activated') }
              )
            }
          >
            Save and activate
          </Button>
        </div>
        <Textarea
          rows={10}
          className="font-mono"
          placeholder={
            'require "fileinto";\n\nif header :contains "subject" "[newsletter]" {\n  fileinto "Newsletters";\n}'
          }
          value={script}
          onChange={(e) => setScript(e.target.value)}
        />
      </CardContent>
    </Card>
  );
}
