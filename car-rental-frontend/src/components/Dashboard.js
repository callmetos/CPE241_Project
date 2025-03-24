import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const Dashboard = () => {
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      setError('You must be logged in to access this page');
      navigate('/login'); // Redirect to login if no token is found
    }
  }, [navigate]);

  return (
    <div>
      <h2>Dashboard</h2>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <p>Welcome to the dashboard! You're logged in.</p>
    </div>
  );
};

export default Dashboard;
