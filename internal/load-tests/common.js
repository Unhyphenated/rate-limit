import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

const baseUrl = 'http://localhost';
const apiPath = '/api/v1/prices';

/** Counts 429 responses (expected under overload). */
export const rateLimited429 = new Counter('rate_limited_429');

export function targetUrl() {
  const path = apiPath.startsWith('/') ? apiPath : `/${apiPath}`;
  return `${baseUrl.replace(/\/$/, '')}${path}`;
}

export function defaultHeaders() {
  const h = {};
  const key = __ENV.API_KEY;
  if (key) {
    h['X-API-KEY'] = key;
  }
  return h;
}

/**
 * GET against the API; treats 200 and 429 as success for threshold/check purposes.
 */
export function hitApi() {
  const res = http.get(targetUrl(), { headers: defaultHeaders() });
  if (res.status === 429) {
    rateLimited429.add(1);
  }
  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
  });
  return res;
}