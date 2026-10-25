export class ApiError extends Error {
  constructor(code, message, violations = []) {
    super(message || code || 'Request failed');
    this.name = 'ApiError';
    this.code = code;
    this.violations = violations;
  }
}

async function request(method, path, body) {
  const options = { method };
  if (body !== undefined) {
    options.headers = { 'Content-Type': 'application/json' };
    options.body = JSON.stringify(body);
  }
  const response = await fetch(path, options);
  if (response.status === 204) {
    return null;
  }
  const text = await response.text();
  let payload = null;
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      payload = null;
    }
  }
  if (!response.ok) {
    const error = payload?.error ?? {};
    throw new ApiError(
      error.code ?? 'http_error',
      error.message ?? `Request failed with status ${response.status}`,
      Array.isArray(error.violations) ? error.violations : [],
    );
  }
  return payload;
}

export const api = {
  get: (path) => request('GET', path),
  post: (path, body = {}) => request('POST', path, body),
  patch: (path, body = {}) => request('PATCH', path, body),
  del: (path) => request('DELETE', path),
};
