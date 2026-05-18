import http from "k6/http";

const BASE_URL = "http://localhost:8081";

const stages = [
    { target: 100, duration: "3m" },
    { target: 300, duration: "3m" },
    { target: 600, duration: "3m" },
    { target: 900, duration: "3m" },
    { target: 1200, duration: "3m" },
];

export const options = {
    scenarios: {
        ramp: {
            executor: "ramping-arrival-rate",
            startRate: stages[0].target,
            timeUnit: "1s",
            preAllocatedVUs: 100,
            maxVUs: 2000,
            stages: stages.map((s) => ({ target: s.target, duration: s.duration })),
        },
    },
    thresholds: {
        http_req_failed: ["rate<0.99"],
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

    http.post(`${BASE_URL}/orders`, payload, {
        headers: { "Content-Type": "application/json" },
    });
}