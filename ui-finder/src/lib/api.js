async function requestJSON(url, options = {}) {
  const response = await fetch(url, options);
  let payload = null;
  let parseError = false;
  try {
    payload = await response.json();
  } catch {
    parseError = true;
  }

  if (!response.ok) {
    const error = payload?.error;
    throw {
      code: error?.code || 'http_error',
      message: error?.message || `Request failed (${response.status})`,
    };
  }

  if (parseError || payload === null) {
    throw { code: 'invalid_response', message: 'Invalid JSON response' };
  }


  return payload;
}

export function postJSON(url, body) {
  return requestJSON(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export function getJSON(url) {
  return requestJSON(url);
}

export function deleteJSON(url) {
  return requestJSON(url, { method: 'DELETE' });
}
