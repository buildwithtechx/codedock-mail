import { toast } from 'sonner';
import { useAuthStore } from '#/stores/auth-store';

export class ApiError extends Error {
  public status: number;
  public data: unknown;

  constructor(status: number, message: string, data?: unknown) {
    super(message);
    this.status = status;
    this.data = data;
    this.name = 'ApiError';
  }
}

async function handleResponseError(response: Response): Promise<never> {
  const isJson = response.headers.get('content-type')?.includes('application/json');
  const data = isJson ? await response.json() : await response.text();
  const errorMessage = data?.message || data?.error || response.statusText || 'An error occurred';
  toast.error(errorMessage);
  throw new ApiError(response.status, errorMessage, data);
}

function prepareHeaders(options?: RequestInit, body?: unknown): Headers {
  const headers = new Headers(options?.headers || {});
  if (!headers.has('Content-Type') && !(body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  const token = useAuthStore.getState().token;
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  return headers;
}

export const apiClient = {
  async fetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const headers = prepareHeaders(options, options.body);
    const response = await fetch(endpoint, { ...options, headers });

    if (response.status === 401) {
      useAuthStore.getState().clearToken();
      window.location.assign('/login');
      throw new ApiError(401, 'Session expired. Please log in again.');
    }

    if (response.status === 204) {
      return undefined as T;
    }

    if (!response.ok) {
      return handleResponseError(response);
    }

    const isJson = response.headers.get('content-type')?.includes('application/json');
    return (isJson ? await response.json() : await response.text()) as T;
  },

  get<T>(endpoint: string, options?: RequestInit) {
    return this.fetch<T>(endpoint, { ...options, method: 'GET' });
  },

  post<T>(endpoint: string, body?: unknown, options?: RequestInit) {
    return this.fetch<T>(endpoint, {
      ...options,
      method: 'POST',
      body: body instanceof FormData ? body : JSON.stringify(body),
    });
  },

  put<T>(endpoint: string, body?: unknown, options?: RequestInit) {
    return this.fetch<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: body instanceof FormData ? body : JSON.stringify(body),
    });
  },

  delete<T>(endpoint: string, options?: RequestInit) {
    return this.fetch<T>(endpoint, { ...options, method: 'DELETE' });
  },
};
