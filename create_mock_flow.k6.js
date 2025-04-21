import http from 'k6/http';
import { check, sleep, fail } from 'k6';

// --- Configuration ---
const API_BASE_URL = 'http://localhost:8080/api'; // Base URL ของ API
const ADMIN_EMAIL = 'admin@carrental.test'; // ต้องสร้าง Admin ไว้ก่อน หรือใช้ Email Admin ที่มีอยู่จริง
const MANAGER_EMAIL = 'manager.mock@carrental.test'; // Email สำหรับ Manager ที่จะสร้าง
const CUSTOMER_EMAIL = 'customer.mock@email.test'; // Email สำหรับ Customer ที่จะสร้าง
const USER_PASSWORD = 'password123'; // รหัสผ่าน (Plain Text) ที่จะใช้

export const options = {
  vus: 1, // ใช้ VU เดียว เพราะต้องทำตามลำดับ
  iterations: 1, // ทำแค่รอบเดียวเพื่อสร้าง 1 flow
  thresholds: {
    // ตั้งค่า thresholds ตามต้องการ หรือเอาออกถ้าไม่ต้องการ check performance
    http_req_failed: ['rate<0.1'], // อนุญาตให้ fail ได้ไม่เกิน 10%
  },
};

// --- Helper Function for Login ---
function login(email, password, userType = 'customer') {
  const loginUrl = `${API_BASE_URL}/auth/${userType}/login`;
  const loginPayload = JSON.stringify({ email: email, password: password });
  const loginParams = { headers: { 'Content-Type': 'application/json' } };
  const loginRes = http.post(loginUrl, loginPayload, loginParams);

  if (!check(loginRes, { [`${userType} login successful (status 200)`]: (r) => r.status === 200 })) {
    fail(`Login failed for ${userType} ${email}: ${loginRes.status} ${loginRes.body}`);
    return null; // Indicate failure
  }
  try {
    return loginRes.json('token'); // Extract token from response JSON
  } catch (e) {
    fail(`Failed to parse token for ${userType} ${email}: ${e}. Response body: ${loginRes.body}`);
    return null;
  }
}

// --- Main k6 Test Logic ---
export default function () {
  let adminToken, managerToken, customerToken;
  let branchId, carId, rentalId;

  // === ขั้นตอนที่ 0: Login Admin (ต้องมี Admin อยู่แล้ว) ===
  console.log('Attempting Admin Login...');
  adminToken = login(ADMIN_EMAIL, USER_PASSWORD, 'employee');
  if (!adminToken) return; // หยุดถ้า Login ไม่ผ่าน
  console.log('Admin Login Successful.');
  sleep(1);

  // === ขั้นตอนที่ 1: สร้าง Branch (ใช้ Admin Token) ===
  console.log('Creating Branch...');
  const branchPayload = JSON.stringify({
    name: `Mock Branch ${Date.now()}`, // Ensure unique name
    address: '1 Mock Address, Mock City',
    phone: '0898765432',
  });
  const createBranchRes = http.post(`${API_BASE_URL}/branches`, branchPayload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${adminToken}`,
    },
  });
  if (!check(createBranchRes, { 'Branch created (status 201)': (r) => r.status === 201 })) {
    fail(`Failed to create branch: ${createBranchRes.status} ${createBranchRes.body}`);
    return;
  }
  branchId = createBranchRes.json('id');
  console.log(`Branch created successfully with ID: ${branchId}`);
  sleep(1);

  // === ขั้นตอนที่ 2: สร้าง Employee (Manager) ใหม่ (Admin Token) ===
  console.log('Creating Manager Employee...');
  const managerPayload = JSON.stringify({
    name: 'Mock Manager',
    email: MANAGER_EMAIL,
    password: USER_PASSWORD, // ส่ง Plain text
    role: 'manager',
  });
  const createManagerRes = http.post(`${API_BASE_URL}/auth/employee/register`, managerPayload, {
    headers: {
      'Content-Type': 'application/json',
      // 'Authorization': `Bearer ${adminToken}`, // Employee registration might be public or require admin
    },
  });
    // Check for 201 Created or 409 Conflict (if already exists from previous run)
  if (!check(createManagerRes, { 'Manager registration submitted (status 201 or 409)': (r) => r.status === 201 || r.status === 409 })) {
      fail(`Failed to register manager: ${createManagerRes.status} ${createManagerRes.body}`);
      return;
  }
  if (createManagerRes.status === 201) {
      console.log('Manager Employee registered successfully.');
  } else {
       console.log('Manager Employee likely already exists (Status 409).');
  }
  sleep(1);

  // === ขั้นตอนที่ 3: Login Manager ที่เพิ่งสร้าง ===
  console.log('Attempting Manager Login...');
  managerToken = login(MANAGER_EMAIL, USER_PASSWORD, 'employee');
  if (!managerToken) return;
  console.log('Manager Login Successful.');
  sleep(1);

  // === ขั้นตอนที่ 4: สร้าง Car (ใช้ Manager Token) ===
  console.log('Creating Car...');
  const carPayload = JSON.stringify({
    brand: 'MockCar',
    model: `Model-${Date.now()}`,
    price_per_day: 1000.00,
    availability: true,
    parking_spot: 'MOCK01',
    branch_id: branchId, // ใช้ branchId ที่ได้จากขั้นตอนที่ 1
    image_url: null,
  });
  const createCarRes = http.post(`${API_BASE_URL}/cars`, carPayload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${managerToken}`, // ใช้ Manager token
    },
  });
  if (!check(createCarRes, { 'Car created (status 201)': (r) => r.status === 201 })) {
    fail(`Failed to create car: ${createCarRes.status} ${createCarRes.body}`);
    return;
  }
  carId = createCarRes.json('id');
  console.log(`Car created successfully with ID: ${carId}`);
  sleep(1);

  // === ขั้นตอนที่ 5: สร้าง Customer ใหม่ ===
  console.log('Creating Customer...');
  const customerPayload = JSON.stringify({
    name: 'Mock Customer',
    email: CUSTOMER_EMAIL,
    phone: '0811223344',
    password: USER_PASSWORD, // ส่ง Plain text
  });
  const createCustomerRes = http.post(`${API_BASE_URL}/auth/customer/register`, customerPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  // Check for 201 Created or 409 Conflict (if already exists)
  if (!check(createCustomerRes, { 'Customer registration submitted (status 201 or 409)': (r) => r.status === 201 || r.status === 409 })) {
      fail(`Failed to register customer: ${createCustomerRes.status} ${createCustomerRes.body}`);
      return;
  }
   if (createCustomerRes.status === 201) {
      console.log('Customer registered successfully.');
  } else {
       console.log('Customer likely already exists (Status 409).');
  }
  sleep(1);

  // === ขั้นตอนที่ 6: Login Customer ที่เพิ่งสร้าง ===
  console.log('Attempting Customer Login...');
  customerToken = login(CUSTOMER_EMAIL, USER_PASSWORD, 'customer');
  if (!customerToken) return;
  console.log('Customer Login Successful.');
  sleep(1);

  // === ขั้นตอนที่ 7: สร้าง Rental (ใช้ Customer Token) ===
  console.log('Creating Rental...');
  const rentalPayload = JSON.stringify({
    car_id: carId, // ใช้ carId ที่ได้จากขั้นตอนที่ 4
    // กำหนดวันเวลา Pickup/Dropoff (ตัวอย่าง)
    pickup_datetime: new Date(Date.now() + 2 * 24 * 60 * 60 * 1000).toISOString(), // อีก 2 วัน
    dropoff_datetime: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString(), // อีก 5 วัน
    pickup_location: null,
  });
  const createRentalRes = http.post(`${API_BASE_URL}/rentals`, rentalPayload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${customerToken}`, // ใช้ Customer token
    },
  });
  if (!check(createRentalRes, { 'Rental created (status 201)': (r) => r.status === 201 })) {
    fail(`Failed to create rental: ${createRentalRes.status} ${createRentalRes.body}`);
    return;
  }
  rentalId = createRentalRes.json('id');
  console.log(`Rental created successfully with ID: ${rentalId}`);
  sleep(1);

  // === ขั้นตอนที่ 8: เปลี่ยนสถานะ Rental เป็น Returned (สมมติว่าเช่าเสร็จแล้ว) (ใช้ Manager Token) ===
  console.log(`Updating Rental ${rentalId} status to Returned...`);
   const returnUrl = `${API_BASE_URL}/rentals/${rentalId}/return`; // Endpoint สำหรับ Return
   const returnRes = http.post(returnUrl, null, { // No body needed for status change usually
     headers: {
       'Authorization': `Bearer ${managerToken}`,
     },
   });
  if (!check(returnRes, { 'Rental status updated to Returned (status 200)': (r) => r.status === 200 })) {
      fail(`Failed to update rental ${rentalId} status: ${returnRes.status} ${returnRes.body}`);
      // ไม่หยุดทำงานทันที อาจจะลองทำขั้นตอนต่อไปได้
      console.log(`Warning: Could not set rental ${rentalId} to Returned.`);
  } else {
      console.log(`Rental ${rentalId} status updated to Returned.`);
  }
  sleep(1);

  // === ขั้นตอนที่ 9: บันทึก Payment (ใช้ Manager Token) ===
  console.log(`Recording Payment for Rental ${rentalId}...`);
  const paymentPayload = JSON.stringify({
    amount: 3000.00, // สมมติราคา 3 วัน * 1000
    payment_status: 'Paid',
    payment_method: 'Cash',
    transaction_id: `mock_txn_${rentalId}`,
  });
  const paymentUrl = `${API_BASE_URL}/rentals/${rentalId}/payments`;
  const paymentRes = http.post(paymentUrl, paymentPayload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${managerToken}`,
    },
  });
    if (!check(paymentRes, { 'Payment recorded (status 201)': (r) => r.status === 201 })) {
      fail(`Failed to record payment for rental ${rentalId}: ${paymentRes.status} ${paymentRes.body}`);
      // ไม่หยุดทำงาน
       console.log(`Warning: Could not record payment for rental ${rentalId}.`);
  } else {
      console.log(`Payment recorded successfully for Rental ${rentalId}.`);
  }
  sleep(1);

  // === ขั้นตอนที่ 10: ส่ง Review (ใช้ Customer Token) ===
  // ต้องมั่นใจว่า Rental อยู่ในสถานะ Returned ก่อน (จากขั้นตอนที่ 8)
   console.log(`Submitting Review for Rental ${rentalId}...`);
   const reviewPayload = JSON.stringify({
       rating: 5,
       comment: `Mock review for rental ${rentalId} by k6`,
   });
   const reviewUrl = `${API_BASE_URL}/rentals/${rentalId}/review`; // Endpoint สำหรับ Submit review
   const reviewRes = http.post(reviewUrl, reviewPayload, {
       headers: {
           'Content-Type': 'application/json',
           'Authorization': `Bearer ${customerToken}`,
       },
   });
   if (!check(reviewRes, { 'Review submitted (status 201)': (r) => r.status === 201 })) {
       fail(`Failed to submit review for rental ${rentalId}: ${reviewRes.status} ${reviewRes.body}`);
       console.log(`Warning: Could not submit review for rental ${rentalId}. Rental might not be 'Returned' yet or other issue.`);
   } else {
       console.log(`Review submitted successfully for Rental ${rentalId}.`);
   }
   sleep(1);

   console.log("=== Mock Data Creation Flow Complete ===");
}