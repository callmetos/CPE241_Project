import React from 'react';
import { Navigate } from 'react-router-dom';

const PrivateRoute = ({ children }) => {
  const token = localStorage.getItem('jwt_token');  // Check if the JWT token exists in localStorage

  if (!token) {
    return <Navigate to="/login" />; // If no token, redirect to login page
  }

  return children;  // Otherwise, render the protected component (e.g., CarRental)
};

export default PrivateRoute;
