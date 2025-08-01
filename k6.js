
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '60s', start:1 , target: 550 }, // ramp-up to 260 VUs over 30 seconds
    ],
};

export default function () {
    const url = 'http://localhost:9999/payments';

    const payload = JSON.stringify({
        correlationId: 'b089cb99-807d-46f2-9659-98bd4eb70542',
        amount: 19.9,
        requestedAt: '2025-07-25T17:06:41.903298842Z',
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    let res = http.post(url, payload, params);

    check(res, {
        'status is 202': (r) => r.status === 202,
    });

    sleep(0.5); // wait between iterations
}