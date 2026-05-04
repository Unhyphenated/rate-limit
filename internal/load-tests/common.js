import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';
import exec from 'k6/execution'

const BASE_URL = __ENV.BASE_URL || 'http://localhost';
const API_PATH = '/api/v1/prices';

export const rateLimited = new Rate('rate_limited_ratio');
export const realErrors = new Rate('real_error_ratio');

export function targetUrl() {
  return `${BASE_URL}${API_PATH}`;
}

export function ipForVU() {
  const vuId = exec.vu.idInTest;
  return `10.0.${Math.floor(vuId / 256)}.${vuId % 256}`
}

export function hitApi(opts = {}) {
  const ip = opts.ip || ipForVU(); 
  const tags = opts.tags || {};

  const res = http.get(targetUrl(), {
    headers: { 'X-Forwarded-For': ip, ...defaultHeaders() },
    tags: { name: 'price_api', ...tags },
    responseCallback: http.expectedStatuses(200, 429),
  });

  rateLimited.add(res.status === 429);
  realErrors.add(res.status !== 200 || res.status !== 429);

  check(res, {
    'valid response': (r) => r.status === 200 || r.status === 429,
    'has rate limit headers': (r) => r.headers['X-Ratelimit-Remaining'] !== undefined;
  });

  return res;
}

function defaultHeaders() {
  const h = {};
  if (__ENV.API_KEY) h['X-API-KEY'] = __ENV.API_KEY;
  return h;
}

