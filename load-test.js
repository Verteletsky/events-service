import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

export const errorRate = new Rate('errors');

export const options = {
    scenarios: {
        ramp_up: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 100 },  // 0 -> 100 VUs за 30 секунд
                { duration: '30s', target: 200 },  // 100 -> 200 VUs за 30 секунд
                { duration: '30s', target: 500 },  // 200 -> 500 VUs за 30 секунд
                { duration: '30s', target: 1000 }, // 500 -> 1000 VUs за 30 секунд
            ],
        }
    },
    thresholds: {
        'http_req_duration': ['p(95)<500'],
        'errors': ['rate<0.1'],
        'http_reqs': ['rate>100'], // Минимальный RPS
    },
};

const BASE_URL = 'http://localhost:8080/v1';

function generateRandomString(length) {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
}

export default function () {
    const eventTypes = ['meeting', 'call', 'task', 'reminder', 'notification'];
    const eventType = eventTypes[Math.floor(Math.random() * eventTypes.length)];
    const uniqueEventType = `${eventType}${generateRandomString(10)}`;

    const startResponse = http.post(`${BASE_URL}/start`, 
        JSON.stringify({ type: uniqueEventType }),
        { headers: {'Content-Type': 'application/json'} }
    );

    check(startResponse, {
        'start event status is 202': (r) => r.status === 202,
    }) || errorRate.add(1);

    sleep(0.1);

    const finishResponse = http.post(`${BASE_URL}/finish`, 
        JSON.stringify({ type: uniqueEventType }),
        { headers: {'Content-Type': 'application/json'} }
    );

    check(finishResponse, {
        'finish event status is 202': (r) => r.status === 202,
    }) || errorRate.add(1);

    sleep(0.1);

    const listResponse = http.get(BASE_URL);

    check(listResponse, {
        'list events status is 200': (r) => r.status === 200,
        'list events has data': (r) => JSON.parse(r.body).events.length > 0,
    }) || errorRate.add(1);
}