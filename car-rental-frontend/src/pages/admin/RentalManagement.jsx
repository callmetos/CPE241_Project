import React, { useState, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAllRentals, updateRentalStatus, deleteRentalAdmin } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import '../../components/admin/AdminCommon.css';

const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const RentalManagement = () => {
  const queryClient = useQueryClient();
  const [error, setError] = useState('');
  const [sortConfig, setSortConfig] = useState({ key: 'id', direction: 'DESC' });
  const [uiFilters, setUiFilters] = useState({ rental_id: '', customer_id: '', car_id: '', status: '' });
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(10);

  const rentalStatuses = ["Pending", "Booked", "Confirmed", "Active", "Returned", "Cancelled", "Pending Verification"];

  const activeQueryParams = useMemo(() => ({
    ...uiFilters,
    rental_id: uiFilters.rental_id ? parseInt(uiFilters.rental_id) || null : null,
    customer_id: uiFilters.customer_id ? parseInt(uiFilters.customer_id) || null : null,
    car_id: uiFilters.car_id ? parseInt(uiFilters.car_id) || null : null,
    page: currentPage,
    limit: itemsPerPage,
    sort_by: sortConfig.key,
    sort_dir: sortConfig.direction.toUpperCase(),
  }), [uiFilters, currentPage, itemsPerPage, sortConfig]);

  const {
    data: rentalsData,
    isLoading: isLoadingRentals,
    isError: isErrorRentals,
    error: rentalsError,
  } = useQuery({
    queryKey: ['rentals', activeQueryParams],
    queryFn: () => fetchAllRentals(activeQueryParams),
    staleTime: 1000 * 60,
    refetchInterval: 1000 * 60 * 2,
    keepPreviousData: true,
  });

  const currentItemsToDisplay = rentalsData?.rentals || [];
  const totalItems = rentalsData?.total_count || 0;
  const totalPages = Math.ceil(totalItems / itemsPerPage);

  const requestSort = useCallback((key) => {
    let direction = 'ASC';
    if (sortConfig.key === key && sortConfig.direction === 'ASC') {
      direction = 'DESC';
    }
    setSortConfig({ key, direction });
    setCurrentPage(1);
  }, [sortConfig]);

  const { mutate: changeRentalStatus, isPending: isUpdatingStatus, variables: updatingVariables } = useMutation({
    mutationFn: ({ rentalId, newStatus }) => updateRentalStatus(rentalId, newStatus),
    onSuccess: () => {
      setError('');
      queryClient.invalidateQueries({ queryKey: ['rentals'] });
      alert(`Rental status updated!`);
    },
    onError: (err) => {
      setError(`Failed to update rental status: ${err.message}`);
    },
  });

  const { mutate: removeRental, isPending: isDeletingRental, variables: deletingRentalId } = useMutation({
    mutationFn: deleteRentalAdmin,
    onSuccess: (data, rentalId) => {
        setError('');
        queryClient.invalidateQueries({ queryKey: ['rentals'] });
        queryClient.invalidateQueries({ queryKey: ['adminDashboardData'] });
        alert(`Rental ID ${rentalId} deleted successfully!`);
    },
    onError: (err, rentalId) => {
        setError(`Failed to delete rental ${rentalId}: ${err.message}`);
        console.error(`Error deleting rental ${rentalId}:`, err);
    },
  });

  const handleStatusChange = (rentalId, currentStatus, targetStatus) => {
    if (window.confirm(`Change rental ${rentalId} status from ${currentStatus} to ${targetStatus}?`)) {
      setError('');
      changeRentalStatus({ rentalId, newStatus: targetStatus });
    }
  };

  const handleDeleteRental = (rentalId) => {
    if (window.confirm(`Are you sure you want to permanently delete Rental ID ${rentalId}? This action cannot be undone.`)) {
        setError('');
        removeRental(rentalId);
    }
  };

  const handleFilterChange = (e) => {
    const { name, value } = e.target;
    setUiFilters(prev => ({ ...prev, [name]: value }));
    setCurrentPage(1);
  };

  const handleClearFilters = () => {
    setUiFilters({ rental_id: '', customer_id: '', car_id: '', status: '' });
    setCurrentPage(1);
  };

  const paginate = (pageNumber) => {
    if (pageNumber > 0 && pageNumber <= totalPages) {
        setCurrentPage(pageNumber);
    }
  };

  const getPossibleActions = (status) => {
    switch (status) {
      case 'Booked':    return ['Confirmed', 'Cancelled'];
      case 'Confirmed': return ['Active', 'Cancelled'];
      case 'Active':    return ['Returned', 'Cancelled'];
      default:          return [];
    }
  };
  const getStatusClassName = (status) => {
     switch (status) {
      case 'Booked': return 'status-booked'; case 'Confirmed': return 'status-confirmed';
      case 'Active': return 'status-active'; case 'Returned': return 'status-returned';
      case 'Cancelled': return 'status-cancelled'; case 'Pending Verification': return 'status-pending-verification';
      default: return '';
    }
  };
  const getActionButtonClassName = (action) => {
    switch (action) {
      case 'Confirmed': return 'admin-button-success'; case 'Active': return 'admin-button-info';
      case 'Returned': return 'admin-button-primary'; case 'Cancelled': return 'admin-button-danger';
      default: return 'admin-button-secondary';
    }
  };

  const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };
  const filterSectionStyle = { display: 'flex', flexWrap: 'wrap', gap: '15px', padding: '15px', marginBottom: '20px', backgroundColor: '#f8f9fa', borderRadius: '4px', border: '1px solid var(--admin-border-color)' };
  const filterGroupStyle = { display: 'flex', flexDirection: 'column', flex: '1 1 150px' };
  const filterLabelStyle = { marginBottom: '5px', fontSize: '0.85em', fontWeight: '500', color: 'var(--admin-text-medium)' };
  const filterInputStyle = { padding: '8px', fontSize: '0.9em', border: '1px solid #ccc', borderRadius: '4px' };
  const filterButtonStyle = { alignSelf: 'flex-end', padding: '8px 15px' };
  const paginationContainerStyle = { display: 'flex', justifyContent: 'center', alignItems: 'center', marginTop: '20px', paddingTop: '15px', borderTop: '1px solid #eee' };
  const paginationButtonStyle = (isActive) => ({ margin: '0 5px', padding: '8px 12px', cursor: 'pointer', backgroundColor: isActive ? 'var(--admin-primary)' : '#f0f0f0', color: isActive ? 'white' : '#333', border: `1px solid ${isActive ? 'var(--admin-primary)' : '#ccc'}`, borderRadius: '4px', fontWeight: isActive ? 'bold' : 'normal' });
  const paginationNavButtonStyle = { ...paginationButtonStyle(false), backgroundColor: '#e9ecef' };

  return (
    <div className="admin-container">
      <div className="admin-header"><h2>Rental Management</h2></div>
      <div style={filterSectionStyle}>
        <div style={filterGroupStyle}><label htmlFor="filter-rental-id" style={filterLabelStyle}>Rental ID</label><input type="number" id="filter-rental-id" name="rental_id" style={filterInputStyle} value={uiFilters.rental_id} onChange={handleFilterChange} placeholder="e.g., 123"/></div>
        <div style={filterGroupStyle}><label htmlFor="filter-customer-id" style={filterLabelStyle}>Customer ID</label><input type="number" id="filter-customer-id" name="customer_id" style={filterInputStyle} value={uiFilters.customer_id} onChange={handleFilterChange} placeholder="e.g., 45"/></div>
        <div style={filterGroupStyle}><label htmlFor="filter-car-id" style={filterLabelStyle}>Car ID</label><input type="number" id="filter-car-id" name="car_id" style={filterInputStyle} value={uiFilters.car_id} onChange={handleFilterChange} placeholder="e.g., 7"/></div>
        <div style={filterGroupStyle}><label htmlFor="filter-status" style={filterLabelStyle}>Status</label><select id="filter-status" name="status" style={filterInputStyle} value={uiFilters.status} onChange={handleFilterChange}><option value="">All Statuses</option>{rentalStatuses.map(status => (<option key={status} value={status}>{status}</option>))}</select></div>
        <div style={{...filterGroupStyle, flexDirection: 'row', alignItems: 'flex-end', gap: '10px', flexBasis: 'auto' }}><button onClick={handleClearFilters} className="admin-button admin-button-secondary" style={filterButtonStyle} disabled={isLoadingRentals}>Clear Filters</button></div>
      </div>
      <ErrorMessage message={error || (isErrorRentals ? `Error: ${rentalsError?.message}` : null)} />
      {isLoadingRentals && <LoadingSpinner />}
      {!isLoadingRentals && (
        <>
        <div className="admin-table-wrapper">
          <table className="admin-table">
            <thead>
              <tr>
                <th style={thSortableStyle} onClick={() => requestSort('id')}>ID <SortIcon direction={sortConfig.key === 'id' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th style={thSortableStyle} onClick={() => requestSort('customer_id')}>Cust. ID <SortIcon direction={sortConfig.key === 'customer_id' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th style={thSortableStyle} onClick={() => requestSort('car_id')}>Car <SortIcon direction={sortConfig.key === 'car_id' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th style={thSortableStyle} onClick={() => requestSort('status')}>Status <SortIcon direction={sortConfig.key === 'status' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th style={thSortableStyle} onClick={() => requestSort('pickup_datetime')}>Pickup <SortIcon direction={sortConfig.key === 'pickup_datetime' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th style={thSortableStyle} onClick={() => requestSort('dropoff_datetime')}>Dropoff <SortIcon direction={sortConfig.key === 'dropoff_datetime' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th className="wrap-text">Pickup Loc.</th>
                <th style={thSortableStyle} onClick={() => requestSort('booking_date')}>Booked <SortIcon direction={sortConfig.key === 'booking_date' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {currentItemsToDisplay.length === 0 ? (
                <tr className="admin-table-placeholder"><td colSpan="9">No rentals found.</td></tr>
              ) : (
                currentItemsToDisplay.map((rental) => {
                  const possibleActions = getPossibleActions(rental.status);
                  const isCurrentlyUpdating = isUpdatingStatus && updatingVariables?.rentalId === rental.id;
                  return (
                    <tr key={rental.id}>
                      <td>{rental.id}</td>
                      <td>{rental.customer_id}</td>
                      <td>{rental.car?.brand} {rental.car?.model} (ID:{rental.car_id})</td>
                      <td><span className={`status-indicator ${getStatusClassName(rental.status)}`}>{rental.status}</span></td>
                      <td>{new Date(rental.pickup_datetime).toLocaleString()}</td>
                      <td>{new Date(rental.dropoff_datetime).toLocaleString()}</td>
                      <td className="wrap-text">{rental.pickup_location || '-'}</td>
                      <td>{rental.booking_date ? new Date(rental.booking_date).toLocaleDateString() : new Date(rental.created_at).toLocaleDateString()}</td>
                      <td className="actions admin-action-buttons">
                        {possibleActions.map(action => (<button key={action} onClick={() => handleStatusChange(rental.id, rental.status, action)} className={`admin-button ${getActionButtonClassName(action)} admin-button-sm`} disabled={isCurrentlyUpdating || (isDeletingRental && deletingRentalId === rental.id)}>{isCurrentlyUpdating && updatingVariables?.rentalId === rental.id && updatingVariables?.newStatus === action ? '...' : action}</button>))}
                        <button
                            onClick={() => handleDeleteRental(rental.id)}
                            className="admin-button admin-button-danger admin-button-sm"
                            disabled={isDeletingRental && deletingRentalId === rental.id || isCurrentlyUpdating}
                            style={{ marginLeft: possibleActions.length > 0 ? '5px' : '0' }}
                        >
                            {isDeletingRental && deletingRentalId === rental.id ? '...' : 'Delete'}
                        </button>
                        {possibleActions.length === 0 && (!isDeletingRental || deletingRentalId !== rental.id) && <span style={{fontSize: '0.85em', color: 'grey', marginLeft: '5px'}}>No status actions</span>}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
        {totalPages > 1 && (
            <div style={paginationContainerStyle}>
                <button onClick={() => paginate(currentPage - 1)} disabled={currentPage === 1} style={paginationNavButtonStyle}>&laquo; Previous</button>
                {Array.from({ length: totalPages }, (_, i) => {
                    const pageNumber = i + 1;
                    const pageRange = 2;
                    const showPage = pageNumber === 1 || pageNumber === totalPages || (pageNumber >= currentPage - pageRange && pageNumber <= currentPage + pageRange);
                    if (showPage) { return (<button key={pageNumber} onClick={() => paginate(pageNumber)} style={paginationButtonStyle(currentPage === pageNumber)}>{pageNumber}</button>); }
                    else if (pageNumber === currentPage - pageRange - 1 || pageNumber === currentPage + pageRange + 1) {
                        if ((pageNumber === 2 && currentPage > 4 && totalPages > 5) || (pageNumber === totalPages - 1 && currentPage < totalPages - 3 && totalPages > 5)) {
                            return <span key={`ellipsis-${pageNumber}`} style={{ margin: '0 5px' }}>...</span>;
                        }
                    }
                    return null;
                })}
                <button onClick={() => paginate(currentPage + 1)} disabled={currentPage === totalPages || totalPages === 0} style={paginationNavButtonStyle}>Next &raquo;</button>
            </div>
        )}
        </>
      )}
    </div>
  );
};
export default RentalManagement;