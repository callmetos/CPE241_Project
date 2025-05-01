import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchDashboardData } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import './AdminDashboard.css'; // Import dashboard specific CSS
import '../../components/admin/AdminCommon.css'; // Import common admin styles

const AdminDashboard = () => {
  const {
    data: dashboardData,
    isLoading,
    isError,
    error
  } = useQuery({
    queryKey: ['adminDashboardData'],
    queryFn: fetchDashboardData,
    staleTime: 1000 * 60 * 5,
  });

  return (
    // Use class from AdminCommon.css
    <div className="admin-container">
      {/* Use class from AdminCommon.css */}
      <div className="admin-header">
        <h2>Dashboard</h2>
        {/* Add buttons or actions here if needed */}
      </div>

      {isLoading && <LoadingSpinner />}
      <ErrorMessage message={isError ? `Error fetching dashboard data: ${error?.message}` : null} />

      {dashboardData && !isLoading && !isError && (
        // Use class from AdminDashboard.css
        <div className="dashboard-grid">
          {/* Use classes from AdminDashboard.css & AdminCommon.css */}
          <div className="admin-card dashboard-card">
            <h3>Total Rentals</h3>
            <p>{dashboardData.total_rentals ?? 'N/A'}</p>
          </div>
          <div className="admin-card dashboard-card">
            <h3>Total Revenue</h3>
            <p>
              {dashboardData.total_revenue !== undefined && dashboardData.total_revenue !== null
                ? `฿${dashboardData.total_revenue.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
                : 'N/A'}
            </p>
          </div>
          <div className="admin-card dashboard-card">
            <h3>Total Customers</h3>
            <p>{dashboardData.total_customers ?? 'N/A'}</p>
          </div>
          {/* Add more dashboard cards here */}
        </div>
      )}
    </div>
  );
};

export default AdminDashboard;
