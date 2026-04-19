import { hitApi } from './common.js';

export const options = {
  discardResponseBodies: true,
  scenarios: {
    steady: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 80,
      maxVUs: 150,
    },
  },
  thresholds: {
    checks: ['rate>0.99'],
  },
};

export default function () {
  hitApi();
}