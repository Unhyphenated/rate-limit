import { hitApi } from './common.js';
import exec from 'k6/execution';

const ATTACKER_IP = '10.99.99.99';

export const options = {
  discardResponseBodies: true,
  scenarios: {
    attacker: {
      executor: 'constant-arrival-rate',
      rate: 1000,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 50,
      maxVUs: 1000,
      exec: 'attackerFn',
      tags: { traffic_type: 'attacker' },
    },
    legitimate: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 30,
      maxVUs: 200,
      exec: 'legitimateFn',
      tags: { traffic_type: 'legitimate' },
    },
  },
  thresholds: {
    // Legitimate users must barely notice the attack
    'http_req_duration{status:200,traffic_type:legitimate}': ['p(99)<30'],
    'rate_limited_ratio{traffic_type:legitimate}': ['rate<0.05'],

    // Attacker must be heavily limited
    'rate_limited_ratio{traffic_type:attacker}': ['rate>0.80'],

    // No real errors anywhere
    'real_error_ratio': ['rate<0.001'],
  },
};

export function attackerFn() {
  hitApi({
    ip: ATTACKER_IP,
    tags: { traffic_type: 'attacker' },
  });
}

export function legitimateFn() {
  hitApi({
    tags: { traffic_type: 'legitimate' },
  });
}