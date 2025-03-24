import React from 'react';
import { Navigate } from 'react-router-dom';

// Protected Route for authentication check
const ProtectedRoute = ({ children }) => {
  const token = localStorage.getItem('token');

  if (!token) {
    // Redirect to login if no token is found
    return <Navigate to="/login" />;
  }

  return children; // Allow access to protected route
};

export default ProtectedRoute;
