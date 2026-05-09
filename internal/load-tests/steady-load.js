import { hitApi } from "./common.js";

export const options = {
    discardResponseBodies: true,
    scenarios: {
        sustained: {
            executor: 'constant-arrival-rate',
            rate: 500,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 100,
            maxVUs: 1000,
            tags: { test: 'sustained' },
        },
    },
    thresholds: {
        'http_req_duration{status:200}': ['p(99)<50'],
        'http_reqs': ['rate>475'],
        'real_error_ratio': ['rate<0.001'],
        'checks': ['rate>0.99'],
    }
};

export default function () {
    hitApi();
}