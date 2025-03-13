import axios from 'axios';

const BASE_URL = "http://localhost:8080/api";

const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  }
});

// Handle JWT Token storage in localStorage (for authentication)
export const setAuthToken = (token) => {
  if (token) {
    api.defaults.headers['Authorization'] = `Bearer ${token}`;
  } else {
    delete api.defaults.headers['Authorization'];
  }
};

// Register employee (for the frontend registration form)
export const register = (name, email, password, role) => {
  return api.post('/register', { name, email, password, role });
};

// Login employee (for the frontend login form)
export const login = (email, password) => {
  return api.post('/login', { email, password });
};

// Fetch available cars
export const getCars = () => {
  return api.get('/cars');
};

export default api;
