import { ApiError } from '#/lib/api-client';

export class MaildError extends Error {
  public status: number;
  public data: unknown;

  constructor(message: string, status: number, data?: unknown) {
    super(message);
    this.name = 'MaildError';
    this.status = status;
    this.data = data;
  }
}

export const handleApiError = (error: unknown): MaildError => {
  if (error instanceof ApiError) {
    let message: string;

    switch (error.status) {
      case 400:
        message = error.message || 'Bad request';
        break;
      case 401:
        message = error.message || 'Unauthorized access';
        break;
      case 403:
        message = error.message || 'Forbidden action';
        break;
      case 404:
        message = error.message || 'Resource not found';
        break;
      case 409:
        message = error.message || 'Already exists';
        break;
      case 422:
        message = error.message || 'Validation error';
        break;
      case 429:
        message = error.message || 'Too many requests. Please try again later';
        break;
      case 500:
        message = error.message || 'Internal server error';
        break;
      default:
        message = error.message || 'An unexpected error occurred';
    }

    return new MaildError(message, error.status, error.data);
  }

  if (error instanceof Error) {
    return new MaildError(error.message, 0);
  }

  return new MaildError('An unexpected error occurred', 0);
};
