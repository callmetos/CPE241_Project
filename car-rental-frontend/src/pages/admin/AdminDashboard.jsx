import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchDashboardData } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import './AdminDashboard.css';
import '../../components/admin/AdminCommon.css';

const AdminDashboard = () => {
  const {
    data: dashboardData,
    isLoading,
    isError,
    error
  } = useQuery({
    queryKey: ['adminDashboardData'],
    queryFn: fetchDashboardData,
    staleTime: 1000 * 60 * 5, // Cache for 5 minutes
  });

  return (
    <div className="admin-container">
      <div className="admin-header">
        <h2>Dashboard</h2>
      </div>

      {isLoading && <LoadingSpinner />}
      <ErrorMessage message={isError ? `Error fetching dashboard data: ${error?.message}` : null} />

      {dashboardData && !isLoading && !isError && (
        <div className="dashboard-grid">
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
          <div className="admin-card dashboard-card">
            <h3>Total Cars</h3>
            <p>{dashboardData.total_cars ?? 'N/A'}</p>
          </div>
          <div className="admin-card dashboard-card">
            <h3 style={{ color: '#28a745' }}>Available Cars</h3>
            <p style={{ color: '#28a745', fontWeight: 'bold' }}>{dashboardData.total_available_cars ?? 'N/A'}</p>
          </div>
          <div className="admin-card dashboard-card">
            <h3 style={{ color: '#dc3545' }}>Unavailable Cars</h3>
            <p style={{ color: '#dc3545', fontWeight: 'bold' }}>{dashboardData.unavailable_cars ?? 'N/A'}</p>
          </div>
           <div className="admin-card dashboard-card">
            <h3>Total Branches</h3>
            <p>{dashboardData.total_branches ?? 'N/A'}</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default AdminDashboard;
