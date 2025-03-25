import axios from 'axios';

const API_URL = 'http://localhost:8080/api'; // Your backend URL

// Register function to interact with the backend API
export const signup = async (email, name, password) => {
  try {
    const response = await axios.post(`${API_URL}/register`, { email,name, password });  // Use /register instead of /signup
    return response.data;  // Returns the token or success response
  } catch (error) {
    console.error('Signup error:', error);
    throw error;  // Rethrow error to be handled by the component
  }
};

// Login function to authenticate and get JWT token
export const login = async (email, password) => {
  try {
    const response = await axios.post(`${API_URL}/login`, { email, password });
    return response.data.token;  // Returns JWT token
  } catch (error) {
    console.error('Login error:', error);
    throw error;  // Rethrow error to be handled by the component
  }
};
