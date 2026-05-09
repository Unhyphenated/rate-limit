import { hitApi } from './common.js';
import exec from 'k6/execution';

export const options = {
  scenarios: {
    phased_test: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 5000,
      stages: [
        { duration: '30s', target: 50 },  // Phase 1
        { duration: '5s', target: 3000 },   // Phase 2
        { duration: '20s', target: 3000 },  // Phase 3
        { duration: '5s', target: 50 },   // Phase 4
        { duration: '30s', target: 50 },   // Phase 5
      ],
    },
  },
  thresholds: {
    'http_req_duration{status:200,phase:phase_1}': ['p(99)<10'],
  
    'http_req_duration{status:200,phase:phase_5}': ['p(99)<15'],
  
    'http_req_duration{status:200,phase:phase_3}': ['p(99)<200'],
  
    'real_error_ratio': ['rate<0.01'],
  
    'http_reqs': ['count>75000'],
  }
};

function getCurrentPhase(elapsedSec) {
  if (elapsedSec <= 30) return 'phase_1';
  if (elapsedSec <= 35) return 'phase_2';
  if (elapsedSec <= 55) return 'phase_3';
  if (elapsedSec <= 60) return 'phase_4';
  return 'phase_5';
}

export default function () {
  const elapsedSec = (Date.now() - exec.scenario.startTime) / 1000;
  const currentPhase = getCurrentPhase(elapsedSec);

  hitApi({
    tags: { phase: currentPhase },
  });
}