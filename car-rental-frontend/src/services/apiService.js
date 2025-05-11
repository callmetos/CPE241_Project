import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
const apiClient = axios.create({ baseURL: API_URL });
export const getToken = () => localStorage.getItem('jwt_token');

apiClient.interceptors.request.use(
  (config) => {
    const token = getToken();
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    if (!(config.data instanceof FormData)) {
        config.headers['Content-Type'] = 'application/json';
    }
    return config;
  },
  (error) => Promise.reject(error)
);

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error(
        'API Response Error:',
        `\n  URL: ${error.config?.method?.toUpperCase()} ${error.config?.url}`,
        `\n  Status: ${error.response?.status}`,
        `\n  Backend Error: ${JSON.stringify(error.response?.data?.error)}`,
        `\n  Axios Message: ${error.message}`
    );
    const message = error.response?.data?.error ||
                    (error.response?.status === 401 ? 'Unauthorized or session expired. Please login again.' : null) ||
                    error.message ||
                    'An unknown API error occurred.';
    const customError = new Error(message);
    customError.response = error.response;
    return Promise.reject(customError);
  }
);

export const signupCustomer = async (email, name, password, phone) => {
  const { data } = await apiClient.post(`/auth/customer/register`, { email, name, password, phone });
  return data;
};
export const loginCustomer = async (email, password) => {
  const { data } = await apiClient.post(`/auth/customer/login`, { email, password });
  if (data?.token) return data.token;
  throw new Error("Token not found in login response.");
};
export const loginEmployee = async (email, password) => {
  const { data } = await apiClient.post(`/auth/employee/login`, { email, password });
  if (data?.token) return data.token;
  throw new Error("Token not found in employee login response.");
};
export const fetchUserProfile = async () => {
  const { data } = await apiClient.get('/me/profile');
  return data;
};
export const fetchAvailableCars = async (params = {}) => {
    const validParams = { availability: true, ...params };
    Object.keys(validParams).forEach(key => (validParams[key] == null || validParams[key] === '') && delete validParams[key]);
    const { data } = await apiClient.get('/cars', { params: validParams });
    if (params.limit === 4 && data && Array.isArray(data.cars)) return data.cars;
    return data || { cars: [], total_count: 0 };
};
export const fetchCarReviews = async (carId) => {
  if (!carId) throw new Error("Car ID is required to fetch reviews.");
  const { data } = await apiClient.get(`/cars/${carId}/reviews`);
  return data || [];
};
export const initiateRentalBooking = async (rentalData) => {
    const payload = {...rentalData, pickup_datetime: new Date(rentalData.pickup_datetime).toISOString(), dropoff_datetime: new Date(rentalData.dropoff_datetime).toISOString()};
    const { data } = await apiClient.post(`/rentals/initiate`, payload);
    if (!data || !data.id) { throw new Error("Backend did not return a rental ID after initiation."); }
    return data;
};
export const fetchRentalDetails = async (rentalId) => {
    if (!rentalId) throw new Error("Rental ID is required.");
    const { data } = await apiClient.get(`/rentals/${rentalId}`);
    return data;
};
export const uploadPaymentSlip = async (rentalId, file) => {
    if (!rentalId || !file) throw new Error("Rental ID and file are required.");
    const formData = new FormData(); formData.append('slip', file);
    const { data } = await apiClient.post(`/rentals/${rentalId}/upload-slip`, formData);
    return data;
};
export const checkPaymentStatus = async (paymentId) => {
    if (!paymentId) throw new Error("Payment ID is required.");
    const { data } = await apiClient.get(`/payments/${paymentId}/status`);
    return data;
};
export const fetchRentalHistory = async (params = {}) => {
    const validParams = { ...params };
    if (validParams.page === undefined) validParams.page = 1;
    if (validParams.limit === undefined) validParams.limit = 10;
    Object.keys(validParams).forEach(key => (validParams[key] == null || validParams[key] === '') && delete validParams[key]);
    const { data } = await apiClient.get('/my/rentals', { params: validParams });
    return data;
};
export const submitReview = async (rentalId, reviewData) => {
    if (!rentalId) throw new Error("Rental ID is required to submit a review.");
    if (!reviewData || !reviewData.rating) throw new Error("Rating is required.");
    const { data } = await apiClient.post(`/rentals/${rentalId}/review`, { rating: reviewData.rating, comment: reviewData.comment || null });
    return data;
};
export const fetchRentalPriceDetails = async (rentalId) => {
    if (!rentalId) throw new Error("Rental ID is required.");
    const { data } = await apiClient.get(`/rentals/${rentalId}/price`);
    return { ...data, currency: data.currency || 'THB' };
};
export const fetchPublicStats = async () => {
  const { data } = await apiClient.get('/public-stats');
  return data;
};
export const fetchDashboardData = async () => {
  const { data } = await apiClient.get('/dashboard');
  return data;
};
export const fetchRevenueReport = async (startDate, endDate) => {
    const { data } = await apiClient.get('/reports/revenue', { params: {start_date: startDate, end_date: endDate} }); return data || [];
};
export const fetchPopularCarsReport = async (limit = 10) => {
    const { data } = await apiClient.get('/reports/popular-cars', { params: { limit } }); return data || [];
};
export const fetchBranchPerformanceReport = async () => {
    const { data } = await apiClient.get('/reports/branch-performance'); return data || [];
};
export const fetchBranches = async () => {
  const { data } = await apiClient.get('/branches');
  return data || [];
};
export const createBranch = async (branchData) => {
  const { data } = await apiClient.post('/branches', { ...branchData, address: branchData.address || null, phone: branchData.phone || null });
  return data;
};
export const updateBranch = async (branchId, branchData) => {
  const { data } = await apiClient.put(`/branches/${branchId}`, { ...branchData, address: branchData.address || null, phone: branchData.phone || null });
  return data;
};
export const deleteBranch = async (branchId) => {
  const { data } = await apiClient.delete(`/branches/${branchId}`);
  return data;
};
export const fetchAllCars = async (params = {}) => {
    const validParams = { ...params };
    if (validParams.availability && typeof validParams.availability === 'string') { validParams.availability = validParams.availability === 'true';}
    Object.keys(validParams).forEach(key => (validParams[key] == null || validParams[key] === '') && delete validParams[key]);
    if (validParams.page === undefined) validParams.page = 1;
    if (validParams.limit === undefined) validParams.limit = 10;
    const { data } = await apiClient.get('/cars', { params: validParams });
    return data;
};
export const createCar = async (carData) => {
    const payload = {...carData, price_per_day: parseFloat(carData.price_per_day) || 0, branch_id: parseInt(carData.branch_id, 10) || 0, availability: carData.availability === undefined ? true : Boolean(carData.availability), parking_spot: carData.parking_spot || null, image_url: carData.image_url || null};
    const { data } = await apiClient.post('/cars', payload);
    return data;
};
export const updateCar = async (carId, carData) => {
    if (!carId) throw new Error("Car ID is required.");
    const payload = {...carData, price_per_day: parseFloat(carData.price_per_day) || 0, branch_id: parseInt(carData.branch_id, 10) || 0, availability: Boolean(carData.availability), parking_spot: carData.parking_spot || null, image_url: carData.image_url || null};
    const { data } = await apiClient.put(`/cars/${carId}`, payload);
    return data;
};
export const deleteCar = async (carId) => {
    if (!carId) throw new Error("Car ID is required.");
    const { data } = await apiClient.delete(`/cars/${carId}`);
    return data;
};
export const fetchAllCustomers = async (params = {}) => {
    const validParams = { ...params };
    Object.keys(validParams).forEach(key => (validParams[key] == null || validParams[key] === '') && delete validParams[key]);
    if (validParams.page === undefined) validParams.page = 1;
    if (validParams.limit === undefined) validParams.limit = 10;
    const { data } = await apiClient.get('/customers', { params: validParams });
    return data;
};
export const updateCustomer = async (customerId, customerData) => {
    if (!customerId) throw new Error("Customer ID is required.");
    const { data } = await apiClient.put(`/customers/${customerId}`, { name: customerData.name, email: customerData.email, phone: customerData.phone || null });
    return data;
};
export const deleteCustomer = async (customerId) => {
    if (!customerId) throw new Error("Customer ID is required.");
    const { data } = await apiClient.delete(`/customers/${customerId}`);
    return data;
};
export const fetchAllRentals = async (params = {}) => {
    const validParams = { ...params };
    Object.keys(validParams).forEach(key => (validParams[key] == null || validParams[key] === '') && delete validParams[key]);
    if (validParams.page === undefined) validParams.page = 1;
    if (validParams.limit === undefined) validParams.limit = 10;
    const { data } = await apiClient.get('/rentals', { params: validParams });
    return data;
};
export const updateRentalStatus = async (rentalId, status) => {
    if (!rentalId || !status) throw new Error("Rental ID and new status are required.");
    let endpoint = '';
    switch (status.toLowerCase()) {
        case 'confirmed': endpoint = `/rentals/${rentalId}/confirm`; break;
        case 'active':    endpoint = `/rentals/${rentalId}/activate`; break;
        case 'returned':  endpoint = `/rentals/${rentalId}/return`; break;
        case 'cancelled': endpoint = `/rentals/${rentalId}/cancel`; break;
        default: throw new Error(`Invalid status for staff update: ${status}`);
    }
    const { data } = await apiClient.post(endpoint);
    return data;
};
export const deleteRentalAdmin = async (rentalId) => {
    if (!rentalId) throw new Error("Rental ID is required for deletion.");
    const { data } = await apiClient.delete(`/rentals/${rentalId}`);
    return data;
};
export const fetchRentalsPendingVerification = async (params = {}) => {
    const { data } = await apiClient.get('/rentals/pending-verification', { params }); return data || [];
};
export const verifyPaymentSlip = async (rentalId, isApproved) => {
    if (!rentalId) throw new Error("Rental ID is required.");
    const { data } = await apiClient.post(`/rentals/${rentalId}/verify-payment`, { approved: Boolean(isApproved) });
    return data;
};
export const fetchAllUsers = async () => {
    const { data } = await apiClient.get('/users'); return data || [];
};
export const createUser = async (userData) => {
    const { data } = await apiClient.post('/users', userData); return data;
};
export const updateUser = async (userId, userData) => {
    if (!userId) throw new Error("User ID is required.");
    const { data } = await apiClient.put(`/users/${userId}`, userData); return data;
};
export const deleteUser = async (userId) => {
    if (!userId) throw new Error("User ID is required.");
    const { data } = await apiClient.delete(`/users/${userId}`); return data;
};
export const fetchAllReviewsAdmin = async (params = {}) => {
    const validParams = { ...params };
    Object.keys(validParams).forEach(key => (validParams[key] == null || validParams[key] === '') && delete validParams[key]);
    if (validParams.page === undefined) validParams.page = 1;
    if (validParams.limit === undefined) validParams.limit = 10;
    const { data } = await apiClient.get('/reviews', { params: validParams });
    return data;
};
export const deleteReview = async (reviewId) => {
    if (!reviewId) throw new Error("Review ID is required for deletion.");
    const { data } = await apiClient.delete(`/reviews/${reviewId}`);
    return data;
};

export default apiClient;