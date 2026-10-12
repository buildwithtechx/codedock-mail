import { useState } from 'react';
import { setToken } from './api';

export function LoginForm({ onDone }: { onDone: () => void }) {
  const [token, setTokenValue] = useState('');
  return (
    <main className="wrap">
      <h1>codedock-mail</h1>
      <p className="muted">
        Enter the maild API token (MAILD_API_TOKEN), or continue without one if the server runs
        open.
      </p>
      <div className="row">
        <input
          type="password"
          placeholder="API token"
          value={token}
          onChange={(e) => setTokenValue(e.target.value)}
        />
        <button
          type="button"
          onClick={() => {
            if (token.trim() !== '') setToken(token.trim());
            onDone();
          }}
        >
          Continue
        </button>
      </div>
    </main>
  );
}
