import React, { useEffect, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';

const Logout = () => {
  const navigate = useNavigate();
  const { logout } = useContext(AuthContext);

  useEffect(() => {
    logout();
    // Redirect to login page after logging out
    navigate('/login', { replace: true });
  }, [logout, navigate]);

  return <p>Logging out...</p>; // Or null, or a spinner
};

export default Logout;