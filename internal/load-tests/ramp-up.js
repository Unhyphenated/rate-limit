import { hitApi } from './common.js';

export const options = {
  discardResponseBodies: true,
  scenarios: {
    ramp: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 400,
      stages: [{ target: 200, duration: '2m' }],
    },
  },
  thresholds: {
    checks: ['rate>0.95'],
  },
};

export default function () {
  hitApi();
}