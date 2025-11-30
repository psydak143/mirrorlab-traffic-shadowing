import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Custom metrics (optional, just for local reporting)
const searchDuration = new Trend('search_duration', true);
const productDuration = new Trend('product_duration', true);
const checkoutDuration = new Trend('checkout_duration', true);
const errors = new Counter('errors');

// Some sample data
const searchQueries = ['ssd', 'ram', 'monitor', 'keyboard', 'mouse', 'hub', ''];
const productIds = ['p-100', 'p-101', 'p-102', 'p-103', 'p-104', 'p-105', 'p-106', 'p-107', 'p-108', 'p-109'];
const emails = ['demo@example.com', 'test@example.com', 'user@example.com'];

// Load profile
export const options = {
  scenarios: {
    shadow_load: {
      executor: 'constant-vus',
      vus: Number(__ENV.VUS) || 50,
      duration: __ENV.DURATION || '5m',
    },
  },
  thresholds: {
    // Global 99th percentile under 500ms (tweak as you like)
    http_req_duration: ['p(99) < 500'],
  },
};

export default function () {
  // Randomly choose which "route" this iteration will hit
  const r = Math.random();
  if (r < 0.5) {
    hitSearch();
  } else if (r < 0.8) {
    hitProduct();
  } else {
    hitCheckout();
  }

  // Small think time so we’re not totally “flat out”
  sleep(0.1);
}

function hitSearch() {
  const q = searchQueries[Math.floor(Math.random() * searchQueries.length)];
  const url = `${BASE_URL}/api/search?q=${encodeURIComponent(q)}`;
  const res = http.get(url, { tags: { route: 'search' } });

  searchDuration.add(res.timings.duration);

  const ok = check(res, {
    'search: status is 2xx': (r) => r.status >= 200 && r.status < 300,
  });
  if (!ok) {
    errors.add(1);
  }
}

function hitProduct() {
  const id = productIds[Math.floor(Math.random() * productIds.length)];
  const url = `${BASE_URL}/api/product/${id}`;
  const res = http.get(url, { tags: { route: 'product' } });

  productDuration.add(res.timings.duration);

  const ok = check(res, {
    'product: status is 2xx or 404': (r) =>
      (r.status >= 200 && r.status < 300) || r.status === 404,
  });
  if (!ok) {
    errors.add(1);
  }
}

function hitCheckout() {
  // simple random cart of 1–3 products
  const itemCount = 1 + Math.floor(Math.random() * 3);
  const ids = [];
  for (let i = 0; i < itemCount; i++) {
    ids.push(productIds[Math.floor(Math.random() * productIds.length)]);
  }
  const email = emails[Math.floor(Math.random() * emails.length)];

  const payload = JSON.stringify({
    productIds: ids,
    email: email,
  });

  const headers = { 'Content-Type': 'application/json' };

  const res = http.post(`${BASE_URL}/api/checkout`, payload, {
    headers,
    tags: { route: 'checkout' },
  });

  checkoutDuration.add(res.timings.duration);

  const ok = check(res, {
    'checkout: status is 2xx or 5xx chaos': (r) =>
      (r.status >= 200 && r.status < 300) || r.status >= 500,
  });
  if (!ok) {
    errors.add(1);
  }
}
