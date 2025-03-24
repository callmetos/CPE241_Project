import axios from 'axios';

// Create an Axios instance to handle API requests
const api = axios.create({
  baseURL: "http://localhost:8080/api", // Your backend API URL
  headers: {
    'Content-Type': 'application/json',
  },
});

// Set JWT token in the request header for authentication
export const setAuthToken = (token) => {
  if (token) {
    api.defaults.headers['Authorization'] = `Bearer ${token}`;
  } else {
    delete api.defaults.headers['Authorization'];
  }
};

// Register an employee
export const register = (name, email, password, role) => {
  return api.post('/register', { name, email, password, role });
};

// Login an employee and store token
export const login = (email, password) => {
  return api.post('/login', { email, password });
};

// Get the list of cars
export const getCars = () => {
  return api.get('/cars');
};

export default api; // Optionally export api instance if you need to use it elsewhere
