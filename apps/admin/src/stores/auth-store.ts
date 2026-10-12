import { create } from 'zustand';

const TOKEN_KEY = 'maild_token';

function readStoredToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(TOKEN_KEY);
}

type AuthState = {
  token: string | null;
  unlocked: boolean;
  setToken: (token: string | null) => void;
  clearToken: () => void;
};

export const useAuthStore = create<AuthState>()((set) => ({
  token: readStoredToken(),
  unlocked: false,
  setToken: (token) => {
    if (typeof window !== 'undefined') {
      if (token) {
        localStorage.setItem(TOKEN_KEY, token);
      } else {
        localStorage.removeItem(TOKEN_KEY);
      }
    }
    set({ token, unlocked: true });
  },
  clearToken: () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem(TOKEN_KEY);
    }
    set({ token: null, unlocked: false });
  },
}));
