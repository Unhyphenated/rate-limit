import { hitApi } from "./common.js";

export const options = {
  discardResponseBodies: true,
  scenarios: {
    burst: {
      executor: "ramping-arrival-rate",
      startRate: 0,
      timeUnit: "1s",
      preAllocatedVUs: 600,
      maxVUs: 1200,
      stages: [
        { target: 500, duration: "1s" },
        { target: 500, duration: "2m" },
      ],
    },
  },
  thresholds: {
    checks: ["rate>0.95"],
  },
};

export default function () {
  hitApi();
}
