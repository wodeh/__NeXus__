// testing/load/scripts/k6-load-test.js
// ============================================================
// K6 LOAD TESTS — High-throughput API testing
// Supports: HTTP/2, WebSocket, gRPC protocols
// ============================================================

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter, Gauge } from 'k6/metrics';
import ws from 'k6/ws';

// Custom metrics
const apiLatency = new Trend('api_latency');
const errorRate = new Rate('errors');
const activeStreams = new Gauge('active_streams');
const iotThroughput = new Counter('iot_messages');

// Test configuration
export const options = {
  scenarios: {
    // Guest portal load
    guest_portal: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '5m', target: 1000 },   // Ramp up
        { duration: '10m', target: 1000 },  // Steady state
        { duration: '5m', target: 2000 },   // Spike
        { duration: '10m', target: 2000 },  // Sustained spike
        { duration: '5m', target: 0 },      // Ramp down
      ],
      gracefulRampDown: '30s',
      exec: 'guestPortal',
    },

    // IPTV streaming load
    iptv_streaming: {
      executor: 'constant-arrival-rate',
      rate: 1000,        // 1000 new streams per minute
      timeUnit: '1m',
      duration: '30m',
      preAllocatedVUs: 500,
      maxVUs: 2000,
      exec: 'iptvStreaming',
    },

    // IoT telemetry ingestion
    iot_telemetry: {
      executor: 'constant-vus',
      vus: 5000,
      duration: '30m',
      exec: 'iotTelemetry',
    },

    // PMS operations
    pms_operations: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '3m', target: 200 },
        { duration: '15m', target: 200 },
        { duration: '2m', target: 500 },
        { duration: '10m', target: 500 },
        { duration: '3m', target: 0 },
      ],
      exec: 'pmsOperations',
    },

    // AI inference load
    ai_inference: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      stages: [
        { duration: '5m', target: 500 },
        { duration: '15m', target: 500 },
        { duration: '5m', target: 1000 },
        { duration: '10m', target: 1000 },
        { duration: '5m', target: 0 },
      ],
      timeUnit: '1m',
      preAllocatedVUs: 300,
      maxVUs: 1000,
      exec: 'aiInference',
    },
  },

  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
    api_latency: ['p(95)<300'],
    errors: ['rate<0.005'],
  },
};

// Setup: Generate test data
export function setup() {
  const tenants = ['hotel-alpha', 'hotel-beta', 'resort-gamma', 'chain-delta'];
  const properties = Array.from({ length: 50 }, (_, i) => `prop_${i + 1}`);
  const rooms = Array.from({ length: 5000 }, (_, i) => `room_${100 + i}`);

  return { tenants, properties, rooms };
}

// Guest Portal Scenario
export function guestPortal(data) {
  group('Guest Portal', () => {
    const tenant = data.tenants[Math.floor(Math.random() * data.tenants.length)];
    const room = data.rooms[Math.floor(Math.random() * data.rooms.length)];

    // Authenticate
    const authRes = http.post(`${__ENV.API_URL}/api/v1/auth/guest`, JSON.stringify({
      room_number: room,
      last_name: 'TestGuest',
    }), {
      headers: { 'Content-Type': 'application/json' },
    });

    check(authRes, {
      'auth status is 200': (r) => r.status === 200,
      'auth response has token': (r) => r.json('token') !== undefined,
    });

    if (authRes.status !== 200) {
      errorRate.add(1);
      return;
    }

    const token = authRes.json('token');
    const headers = {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      'X-Tenant-ID': tenant,
    };

    // Browse services
    const services = [
      '/api/v1/guest/services',
      '/api/v1/guest/dining/menu',
      '/api/v1/guest/spa/services',
      '/api/v1/guest/activities',
    ];

    const serviceRes = http.get(
      `${__ENV.API_URL}${services[Math.floor(Math.random() * services.length)]}`,
      { headers }
    );

    check(serviceRes, {
      'service response is 200': (r) => r.status === 200,
    });

    apiLatency.add(serviceRes.timings.duration);

    // AI Concierge interaction
    if (Math.random() < 0.3) {
      const conciergeRes = http.post(`${__ENV.API_URL}/api/v1/ai/concierge`, JSON.stringify({
        prompt: 'What time is checkout?',
        language: 'en',
      }), { headers });

      check(conciergeRes, {
        'concierge response is 200': (r) => r.status === 200,
        'concierge has content': (r) => r.json('content') !== undefined,
      });

      apiLatency.add(conciergeRes.timings.duration);
    }

    sleep(Math.random() * 5 + 2);
  });
}

// IPTV Streaming Scenario
export function iptvStreaming(data) {
  group('IPTV Streaming', () => {
    const room = data.rooms[Math.floor(Math.random() * data.rooms.length)];
    const channel = Math.floor(Math.random() * 500) + 1;
    const profile = ['1080p', '720p', '480p', '360p'][Math.floor(Math.random() * 4)];

    // Authenticate stream
    const authRes = http.post(`${__ENV.STREAMING_URL}/api/v1/iptv/auth`, JSON.stringify({
      room_id: room,
      device_id: `tv_${Math.floor(Math.random() * 10000)}`,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });

    if (authRes.status !== 200) {
      errorRate.add(1);
      return;
    }

    const streamToken = authRes.json('token');

    // Request HLS manifest
    const manifestRes = http.get(
      `${__ENV.STREAMING_URL}/hls/channel_${channel}/${profile}/master.m3u8`,
      { headers: { 'X-Stream-Token': streamToken } }
    );

    check(manifestRes, {
      'manifest is 200': (r) => r.status === 200,
      'manifest has playlists': (r) => r.body.includes('#EXT-X-STREAM-INF'),
    });

    activeStreams.add(1);

    // Simulate segment downloads
    const segments = Math.floor(Math.random() * 60) + 10;
    for (let i = 0; i < segments; i++) {
      const segmentRes = http.get(
        `${__ENV.STREAMING_URL}/hls/channel_${channel}/${profile}/segment_${i}.ts`,
        { headers: { 'X-Stream-Token': streamToken } }
      );

      check(segmentRes, {
        'segment is 200': (r) => r.status === 200,
      });

      sleep(6); // 6-second segments
    }

    activeStreams.add(-1);
  });
}

// IoT Telemetry Scenario
export function iotTelemetry(data) {
  group('IoT Telemetry', () => {
    const property = data.properties[Math.floor(Math.random() * data.properties.length)];
    const deviceTypes = ['thermostat', 'occupancy_sensor', 'light', 'lock', 'curtain'];

    // Send individual telemetry
    const telemetry = {
      device_id: `device_${Math.floor(Math.random() * 100000)}`,
      property_id: property,
      timestamp: new Date().toISOString(),
      type: deviceTypes[Math.floor(Math.random() * deviceTypes.length)],
      data: {
        temperature: 18 + Math.random() * 8,
        humidity: 30 + Math.random() * 40,
        occupied: Math.random() > 0.5,
        brightness: Math.floor(Math.random() * 100),
      },
    };

    const res = http.post(`${__ENV.API_URL}/api/v1/iot/telemetry`, JSON.stringify(telemetry), {
      headers: { 'Content-Type': 'application/json' },
    });

    check(res, {
      'telemetry accepted': (r) => r.status === 202 || r.status === 200,
    });

    iotThroughput.add(1);

    // Batch telemetry (every 10th request)
    if (Math.random() < 0.1) {
      const batch = Array.from({ length: 100 }, () => ({
        device_id: `device_${Math.floor(Math.random() * 100000)}`,
        property_id: property,
        timestamp: new Date().toISOString(),
        type: deviceTypes[Math.floor(Math.random() * deviceTypes.length)],
        data: telemetry.data,
      }));

      const batchRes = http.post(
        `${__ENV.API_URL}/api/v1/iot/telemetry/batch`,
        JSON.stringify({ devices: batch }),
        { headers: { 'Content-Type': 'application/json' } }
      );

      check(batchRes, {
        'batch telemetry accepted': (r) => r.status === 202,
      });

      iotThroughput.add(100);
    }

    sleep(Math.random() * 2);
  });
}

// PMS Operations Scenario
export function pmsOperations(data) {
  group('PMS Operations', () => {
    const tenant = data.tenants[Math.floor(Math.random() * data.tenants.length)];

    // Staff authentication
    const authRes = http.post(`${__ENV.API_URL}/api/v1/auth/staff`, JSON.stringify({
      employee_id: `emp_${Math.floor(Math.random() * 1000)}`,
      pin: '1234',
    }), {
      headers: { 'Content-Type': 'application/json' },
    });

    if (authRes.status !== 200) {
      errorRate.add(1);
      return;
    }

    const token = authRes.json('token');
    const headers = {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      'X-Tenant-ID': tenant,
    };

    // Random PMS operation
    const operations = [
      () => checkInGuest(headers, data),
      () => updateHousekeeping(headers, data),
      () => processPayment(headers, data),
      () => createMaintenanceRequest(headers, data),
      () => updateInventory(headers, data),
    ];

    operations[Math.floor(Math.random() * operations.length)]();

    sleep(Math.random() * 3 + 1);
  });
}

function checkInGuest(headers, data) {
  const room = data.rooms[Math.floor(Math.random() * data.rooms.length)];

  const res = http.post(`${__ENV.API_URL}/api/v1/reservations/checkin`, JSON.stringify({
    reservation_id: `res_${Math.floor(Math.random() * 1000000)}`,
    room_id: room,
    guest_preferences: ['high_floor', 'non_smoking'],
  }), { headers });

  check(res, {
    'check-in successful': (r) => r.status === 200,
  });

  apiLatency.add(res.timings.duration);
}

function updateHousekeeping(headers, data) {
  const room = data.rooms[Math.floor(Math.random() * data.rooms.length)];

  const res = http.patch(`${__ENV.API_URL}/api/v1/housekeeping/rooms/${room}`, JSON.stringify({
    status: ['clean', 'dirty', 'inspected'][Math.floor(Math.random() * 3)],
    completed_by: `hk_${Math.floor(Math.random() * 50)}`,
  }), { headers });

  check(res, {
    'housekeeping updated': (r) => r.status === 200,
  });

  apiLatency.add(res.timings.duration);
}

function processPayment(headers, data) {
  const folioId = `folio_${Math.floor(Math.random() * 1000000)}`;

  const res = http.post(`${__ENV.API_URL}/api/v1/folios/${folioId}/charges`, JSON.stringify({
    category: ['room', 'f&b', 'spa'][Math.floor(Math.random() * 3)],
    amount: 10 + Math.random() * 490,
    description: 'Guest service',
  }), { headers });

  check(res, {
    'charge processed': (r) => r.status === 201,
  });

  apiLatency.add(res.timings.duration);
}

function createMaintenanceRequest(headers, data) {
  const room = data.rooms[Math.floor(Math.random() * data.rooms.length)];

  const res = http.post(`${__ENV.API_URL}/api/v1/maintenance/requests`, JSON.stringify({
    room_id: room,
    category: ['plumbing', 'electrical', 'hvac'][Math.floor(Math.random() * 3)],
    priority: ['low', 'medium', 'high'][Math.floor(Math.random() * 3)],
    description: 'Guest reported issue',
  }), { headers });

  check(res, {
    'maintenance request created': (r) => r.status === 201,
  });

  apiLatency.add(res.timings.duration);
}

function updateInventory(headers, data) {
  const res = http.patch(`${__ENV.API_URL}/api/v1/inventory/items/${Math.floor(Math.random() * 10000)}`, JSON.stringify({
    quantity_change: Math.floor(Math.random() * 20) - 10,
    reason: 'consumption',
  }), { headers });

  check(res, {
    'inventory updated': (r) => r.status === 200,
  });

  apiLatency.add(res.timings.duration);
}

// AI Inference Scenario
export function aiInference(data) {
  group('AI Inference', () => {
    const tenant = data.tenants[Math.floor(Math.random() * data.tenants.length)];

    const inferenceTypes = [
      { endpoint: '/api/v1/ai/concierge', payload: { prompt: 'What time is checkout?', use_rag: true } },
      { endpoint: '/api/v1/ai/sentiment', payload: { text: 'Great hotel, loved the service!' } },
      { endpoint: '/api/v1/ai/forecast/occupancy', payload: { property_id: 'prop_1', horizon_days: 30 } },
      { endpoint: '/api/v1/ai/recommendations', payload: { guest_id: 'guest_123', context: { time_of_day: 'evening' } } },
    ];

    const inference = inferenceTypes[Math.floor(Math.random() * inferenceTypes.length)];

    const res = http.post(`${__ENV.API_URL}${inference.endpoint}`, JSON.stringify({
      ...inference.payload,
      tenant_id: tenant,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });

    check(res, {
      'inference successful': (r) => r.status === 200,
      'inference has result': (r) => r.json('content') !== undefined || r.json('result') !== undefined,
    });

    apiLatency.add(res.timings.duration);

    sleep(Math.random() * 2);
  });
}

// Teardown: Cleanup
export function teardown(data) {
  console.log('Load test completed');
  console.log(`Tested ${data.rooms.length} rooms across ${data.properties.length} properties`);
}
