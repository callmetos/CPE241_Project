// src/services/apiService.js
import axios from 'axios';
import { getToken } from './authService'; // Use helper to get token

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

// Central Axios instance
const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request Interceptor to add token
apiClient.interceptors.request.use(
  (config) => {
    const token = getToken();
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response Interceptor to handle errors globally
apiClient.interceptors.response.use(
  (response) => response, // Pass through successful responses
  (error) => {
    console.error('API Response Error Interceptor:', error.config?.url, error.response?.status, error.message);
    const message = error.response?.data?.error || // Use backend error message if available
                    (error.response?.status === 401 ? 'Unauthorized or session expired. Please login again.' :
                     error.message) || // Fallback to generic Axios error message
                     'An unknown API error occurred.'; // Ultimate fallback

    // Create a new error object with a better message
    const customError = new Error(message);
    customError.response = error.response; // Attach original response if available
    customError.request = error.request; // Attach request if available
    customError.config = error.config; // Attach config if available

    // Special handling for 401 could happen here (e.g., trigger logout)
    // if (error.response?.status === 401) {
    //    // Maybe call logout from context here, but be careful with circular dependencies
    // }

    return Promise.reject(customError); // Reject with the enhanced error object
  }
);


// --- Specific API functions using the apiClient ---

/**
 * Fetches available cars based on optional criteria.
 * @param {object} [criteria] - Filtering criteria (e.g., { type: 'short-term' }). Backend must support these.
 * @returns {Promise<Array>} Array of car objects.
 */
export const fetchAvailableCars = async (criteria) => {
    const params = { availability: true, ...criteria }; // Add criteria to params
    // Note: Backend needs to be updated to handle criteria like 'type'
    console.log("Fetching cars with params:", params);
    const { data } = await apiClient.get('/cars', { params });
    return data || []; // Ensure returning an array
};

/**
 * Fetches the profile of the currently logged-in user.
 * @returns {Promise<object>} User profile object.
 */
export const fetchUserProfile = async () => {
    const { data } = await apiClient.get('/me/profile');
    return data;
};

/**
 * Fetches the rental history for the currently logged-in user.
 * @returns {Promise<Array>} Array of rental objects.
 */
export const fetchRentalHistory = async () => {
    const { data } = await apiClient.get('/my/rentals');
    return data || [];
};

/**
 * Creates a new rental booking.
 * @param {object} rentalData - Data for the new rental (e.g., { car_id, pickup_datetime, dropoff_datetime }).
 * @returns {Promise<object>} The newly created rental object.
 */
export const createRentalBooking = async (rentalData) => {
    if (!rentalData.pickup_datetime || !rentalData.dropoff_datetime) {
       throw new Error("Pickup and Dropoff times are required.");
    }
     // Ensure dates are ISO strings before sending
     rentalData.pickup_datetime = new Date(rentalData.pickup_datetime).toISOString();
     rentalData.dropoff_datetime = new Date(rentalData.dropoff_datetime).toISOString();

    const { data } = await apiClient.post('/rentals', rentalData);
    return data;
};

// --- Add other API functions as needed ---
// export const cancelRental = async (rentalId) => apiClient.post(`/my/rentals/${rentalId}/cancel`);
// export const submitReview = async (rentalId, reviewData) => apiClient.post(`/my/rentals/${rentalId}/review`, reviewData);
// export const fetchBranches = async () => apiClient.get('/branches');


export default apiClient; // Export the instance for potential direct use