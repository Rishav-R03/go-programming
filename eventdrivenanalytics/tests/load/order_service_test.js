import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom Metrics
const ordersCreatedCounter = new Counter('orders_created_total');
const errorRate = new Rate('errors');
const orderCreationTrend = new Trend('order_creation_duration');

export const options = {
  stages: [
    { duration: '30s', target: 20 },  // Ramp up to 20 VUs
    { duration: '1m', target: 50 },   // Sustained load at 50 VUs
    { duration: '30s', target: 100 }, // Peak load at 100 VUs
    { duration: '30s', target: 0 },   // Ramp down to 0 VUs
  ],
  thresholds: {
    http_req_duration: ['p(95)<200', 'p(99)<500'], // SLA: 95% < 200ms, 99% < 500ms
    http_req_failed: ['rate<0.01'],                 // Error rate < 1%
    errors: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.ORDER_SERVICE_URL || 'http://localhost:8000';

export default function () {
  // Constrained to match seed data (2 customers, 2 restaurants, 3 menu items)
  const payload = JSON.stringify({
    customer_id: Math.floor(Math.random() * 2) + 1,   // IDs 1 or 2
    restaurant_id: Math.floor(Math.random() * 2) + 1, // IDs 1 or 2
    item_id: Math.floor(Math.random() * 3) + 1,       // IDs 1, 2, or 3
    quantity: Math.floor(Math.random() * 5) + 1,
    price: 199.00
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(`${BASE_URL}/orders`, payload, params);

  const success = check(res, {
    'status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'has valid response body': (r) => r.body && r.body.length > 0,
  });

  errorRate.add(!success);
  orderCreationTrend.add(res.timings.duration);

  if (success) {
    ordersCreatedCounter.add(1);
  }

  sleep(0.1 + Math.random() * 0.4);
}