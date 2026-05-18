import http from "k6/http";
import { Counter } from "k6/metrics";
import { check } from "k6";

const BASE_URL = "http://localhost:8081";

const accepted = new Counter("orders_created_total");
const rejected503 = new Counter("orders_rejected_503_total");
const badRequest400 = new Counter("orders_bad_request_400_total");
const otherResponses = new Counter("orders_other_responses_total");

const stages = [
    { target: 20, duration: "2m" },
    { target: 50, duration: "2m" },
    { target: 100, duration: "2m" },
    { target: 200, duration: "2m" },
    { target: 350, duration: "2m" },
    { target: 500, duration: "2m" },
];

export const options = {
    scenarios: {
        ramp: {
            executor: "ramping-arrival-rate",
            startRate: stages[0].target,
            timeUnit: "1s",
            preAllocatedVUs: 200,
            maxVUs: 3000,
            stages: stages.map((s) => ({ target: s.target, duration: s.duration })),
        },
    },
};

const PRODUCTS = [
    { id: "101", name: "Wireless Mouse" },
    { id: "102", name: "Mechanical Keyboard" },
    { id: "103", name: "USB-C Hub" },
    { id: "104", name: "4K Monitor" },
    { id: "105", name: "Noise Cancelling Headphones" },
    { id: "106", name: "Laptop Stand" },
    { id: "107", name: "Webcam" },
    { id: "108", name: "Microphone" },
];

export default function () {
    const itemCount = Math.floor(Math.random() * 3) + 1;
    const items = Array.from({ length: itemCount }, () => {
        const p = PRODUCTS[Math.floor(Math.random() * PRODUCTS.length)];
        return {
            product_id: p.id,
            name: p.name,
            quantity: Math.floor(Math.random() * 5) + 1,
            price: parseFloat((Math.random() * 990 + 10).toFixed(2)),
        };
    });

    const payload = JSON.stringify({
        user_id: String(Math.floor(Math.random() * 100000) + 1),
        items,
    });

    const res = http.post(`${BASE_URL}/orders`, payload, {
        headers: { "Content-Type": "application/json" },
    });

    if (res.status === 201) {
        accepted.add(1);
    } else if (res.status === 503) {
        rejected503.add(1);
    } else if (res.status === 400) {
        badRequest400.add(1);
    } else {
        otherResponses.add(1);
    }

    check(res, {
        "status is 201 or 503": (r) => r.status === 201 || r.status === 503,
    });
}