// src/services/authService.js
import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

/**
 * Attempts to register a new customer.
 * @param {string} email
 * @param {string} name
 * @param {string} password
 * @returns {Promise<object>} Backend response (e.g., { message, customer })
 * @throws {Error} If registration fails, with message from backend or generic message.
 */
export const signup = async (email, name, password) => {
     try {
        const response = await axios.post(`${API_URL}/customer/register`, { email, name, password });
        return response.data;
      } catch (error) {
        console.error('Signup Service Error:', error.response ? error.response.data : error.message);
        throw new Error(error.response?.data?.error || 'Signup failed. Please try again.');
      }
};

/**
 * Attempts to log in a customer.
 * @param {string} email
 * @param {string} password
 * @returns {Promise<string>} The JWT token upon successful login.
 * @throws {Error} If login fails, with message from backend or generic message.
 */
export const login = async (email, password) => {
  try {
    const response = await axios.post(`${API_URL}/customer/login`, { email, password });
    if (response.data && response.data.token) {
      return response.data.token; // Return only the token
    } else {
      throw new Error("Token not found in login response.");
    }
  } catch (error) {
    console.error('Login Service Error:', error.response ? error.response.data : error.message);
    throw new Error(error.response?.data?.error || 'Invalid email or password.');
  }
};

// Helper function to get token from localStorage (can be used by apiService)
export const getToken = () => {
  return localStorage.getItem('jwt_token');
};