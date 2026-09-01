// Phase 5 load test using k6 (https://k6.io).
// Install: see https://grafana.com/docs/k6/latest/set-up/install-k6/
//
// Usage:
//   k6 run scripts/loadtest_k6.js
//   k6 run --vus 50 --duration 30s scripts/loadtest_k6.js
//
// Each virtual user gets its own API key so you can see the per-client
// sliding window kick in independently (VU 3 hitting 429 shouldn't affect
// VU 7's quota) — a good sanity check the Redis key partitioning works.
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const HOST = __ENV.HOST || 'http://localhost:8000';

const allowed = new Counter('rate_limit_allowed');
const blocked = new Counter('rate_limit_blocked');

export const options = {
  scenarios: {
    steady_load: {
      executor: 'constant-arrival-rate',
      rate: Number(__ENV.RATE || 200),      // requests per timeUnit
      timeUnit: '1s',
      duration: __ENV.DURATION || '30s',
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<100', 'p(99)<250'], // ms — tune to your SLO
  },
};

export default function () {
  const clientID = `k6-vu-${__VU}`;
  const res = http.get(`${HOST}/check`, {
    headers: { 'X-API-KEY': clientID },
  });

  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'has X-RateLimit-Remaining header': (r) => r.headers['X-Ratelimit-Remaining'] !== undefined,
  });

  if (res.status === 200) {
    allowed.add(1);
  } else if (res.status === 429) {
    blocked.add(1);
  }

  sleep(0.01);
}
