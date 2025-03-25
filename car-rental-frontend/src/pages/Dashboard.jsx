import React, { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';

const Dashboard = () => {
  const [dashboardData, setDashboardData] = useState(null);
  const [error, setError] = useState('');

  // Fetch dashboard data from backend (e.g., rental statistics)
  useEffect(() => {
    const token = localStorage.getItem('jwt_token');

    if (!token) {
      setError('You are not logged in');
      return;
    }

    const fetchDashboardData = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/dashboard', {
          headers: {
            Authorization: `Bearer ${token}`, // Send JWT token in headers
          },
        });
        const data = await response.json();
        setDashboardData(data);
      } catch (err) {
        setError('Failed to fetch dashboard data');
      }
    };

    fetchDashboardData();
  }, []);

  if (error) {
    return <Navigate to="/login" />;
  }

  return (
    <div className="dashboard-container">
      <h2>Dashboard</h2>
      {dashboardData ? (
        <div>
          <h3>Total Rentals: {dashboardData.total_rentals}</h3>
          <h3>Total Revenue: {dashboardData.total_revenue}</h3>
          <h3>Total Customers: {dashboardData.total_customers}</h3>
        </div>
      ) : (
        <p>Loading dashboard data...</p>
      )}
    </div>
  );
};

export default Dashboard;
