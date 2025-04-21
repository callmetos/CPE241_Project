import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

// --- Configuration ---
const API_BASE_URL = 'http://localhost:8080/api/auth'; // เปลี่ยน URL ถ้าจำเป็น
const NUM_CUSTOMERS = 80; // จำนวนลูกค้าที่ต้องการสร้าง
const VUS = 10; // จำนวน Virtual Users ที่จะรันพร้อมกัน (ปรับได้)
// -------------------

export const options = {
  vus: VUS,
  iterations: NUM_CUSTOMERS, // ให้ k6 รันทั้งหมด NUM_CUSTOMERS ครั้ง
  thresholds: {
    http_req_failed: ['rate<0.01'], // http errors should be less than 1%
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
  },
};

// Function to generate random phone number (like SQL)
function generatePhone() {
  const randomDigits = Math.floor(Math.random() * 100000000).toString().padStart(8, '0');
  return `08${randomDigits}`;
}

export default function () {
  // Generate unique data based on VU id and iteration number
  const vuID = __VU; // Virtual User ID (starts from 1)
  const iter = __ITER; // Iteration number (starts from 0)
  const uniqueSuffix = `${vuID}-${iter}`; // Create a unique combination

  const customerName = `ลูกค้าทดสอบ ${uniqueSuffix}`;
  const customerEmail = `customer_${uniqueSuffix}@email.test`;
  const customerPhone = generatePhone();
  const customerPassword = 'password123'; // Using plain text as requested

  // --- API Request ---
  const url = `${API_BASE_URL}/customer/register`;
  const payload = JSON.stringify({
    name: customerName,
    email: customerEmail,
    phone: customerPhone, // Can be null if needed: Math.random() > 0.5 ? customerPhone : null,
    password: customerPassword,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  // Send POST request
  const res = http.post(url, payload, params);

  // --- Check Response ---
  const checkRes = check(res, {
    'status is 201 Created': (r) => r.status === 201,
    'response body contains message': (r) => r.body.includes('Registration successful!'),
  });

  if (!checkRes) {
    console.error(`Failed registration for ${customerEmail}: ${res.status} - ${res.body}`);
  }

  // Add a small sleep between requests (e.g., 1 second)
  sleep(1);
}