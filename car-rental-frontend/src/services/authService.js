import axios from 'axios';
import apiClient from './apiService';


export const signup = async (email, name, password, phone) => {
     try {
        const response = await apiClient.post(`/auth/customer/register`, { email, name, password, phone });
        return response.data;
      } catch (error) {

        throw new Error(error.response?.data?.error || error.message || 'Signup failed. Please try again.');
      }
};


export const login = async (email, password) => {
  try {

    const response = await apiClient.post(`/auth/customer/login`, { email, password });
    if (response.data && response.data.token) {
      return response.data.token;
    } else {
      throw new Error("Token not found in login response.");
    }
  } catch (error) {

    throw new Error(error.response?.data?.error || error.message || 'Invalid email or password.');
  }
};


export const getToken = () => {
  return localStorage.getItem('jwt_token');
};