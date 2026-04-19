import { hitApi } from './common.js';

const duration = '2m';

export const options = {
  discardResponseBodies: true,
  scenarios: {
    hot_client: {
      executor: 'constant-arrival-rate',
      rate: 1000,
      timeUnit: '1s',
      duration,
      preAllocatedVUs: 400,
      maxVUs: 2500,
    },
  },
  thresholds: {
    checks: ['rate>0.90'],
  },
};

export default function () {
  hitApi();
}