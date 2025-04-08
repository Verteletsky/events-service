import http from 'k6/http';
import {check, sleep} from 'k6';
import {randomString} from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
    stages: [
        {duration: '30s', target: 20}, // Разогрев
        {duration: '1m', target: 50},  // Нормальная нагрузка
        {duration: '30s', target: 100}, // Пиковая нагрузка
        {duration: '30s', target: 0},   // Снижение нагрузки
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = 'http://localhost:8080/v1';

export default function () {
    const eventTypes = ['meeting', 'call', 'task', 'reminder', 'notification'];
    const eventType = eventTypes[Math.floor(Math.random() * eventTypes.length)];

    const startResponse = http.post(`${BASE_URL}/start`, JSON.stringify({
        type: eventType,
        payload: {
            title: randomString(10),
            description: randomString(50),
        }
    }), {
        headers: {'Content-Type': 'application/json'},
    });

    check(startResponse, {
        'start event status is 200': (r) => r.status === 200,
    });

    sleep(Math.random() * 2);

    const finishResponse = http.post(`${BASE_URL}/finish`, JSON.stringify({
        type: eventType,
    }), {
        headers: {'Content-Type': 'application/json'},
    });

    check(finishResponse, {
        'finish event status is 200': (r) => r.status === 200,
    });

    const listResponse = http.get(BASE_URL);

    check(listResponse, {
        'list events status is 200': (r) => r.status === 200,
        'list events has data': (r) => JSON.parse(r.body).events.length > 0,
    });

    sleep(Math.random() * 1);
} 