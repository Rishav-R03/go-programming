import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom Metrics
const errorRate = new Rate('errors');
const dashboardQueryTrend = new Trend('dashboard_query_duration');

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 VUs
    { duration: '1m', target: 30 },   // Sustained read load at 30 VUs
    { duration: '30s', target: 50 },  // Peak read load at 50 VUs
    { duration: '30s', target: 0 },   // Ramp down to 0 VUs
  ],
  thresholds: {
    http_req_duration: ['p(95)<150', 'p(99)<300'], // OLAP query SLA: 95% < 150ms
    http_req_failed: ['rate<0.01'],                 // Error rate < 1%
    errors: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.ANALYTICS_SERVICE_URL || 'http://localhost:8081';

export default function () {
  // FIXED: Restrict to valid seeded restaurants (IDs 1 and 2)
  const restaurantId = Math.floor(Math.random() * 2) + 1;

  const res = http.get(`${BASE_URL}/analytics/dashboard?restaurant_id=${restaurantId}`);

  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'dashboard contains data': (r) => r.body && r.body.length > 0,
  });

  errorRate.add(!success);
  dashboardQueryTrend.add(res.timings.duration);

  sleep(0.2 + Math.random() * 0.5); // Thinking time: 200ms - 700ms
}