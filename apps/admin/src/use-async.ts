import { type DependencyList, useEffect, useRef, useState } from 'react';

export function useAsync<T>(
  run: () => Promise<T>,
  deps: DependencyList
): {
  data: T | null;
  error: string | null;
  reload: () => void;
} {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [nonce, setNonce] = useState(0);
  const runRef = useRef(run);
  runRef.current = run;
  // biome-ignore lint/correctness/useExhaustiveDependencies: nonce is a manual reload signal, not a reactive value
  useEffect(() => {
    let live = true;
    runRef
      .current()
      .then((result) => {
        if (live) {
          setData(result);
          setError(null);
        }
      })
      .catch((err: unknown) => {
        if (live) setError(err instanceof Error ? err.message : 'failed');
      });
    return () => {
      live = false;
    };
  }, [...deps, nonce]);
  return { data, error, reload: () => setNonce((n) => n + 1) };
}
