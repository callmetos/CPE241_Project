import React, { useState, useMemo, useCallback } from 'react'; // Import hooks
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAllRentals, updateRentalStatus } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import '../../components/admin/AdminCommon.css'; // Import common styles

// --- Sort Icon Component (Optional) ---
const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};


const RentalManagement = () => {
  const queryClient = useQueryClient();
  const [error, setError] = useState('');
  // --- State for sorting ---
  const [sortConfig, setSortConfig] = useState({ key: 'id', direction: 'ascending' }); // Default sort

  // Fetch rentals data
  const {
    data: rentals = [],
    isLoading: isLoadingRentals,
    isError: isErrorRentals,
    error: rentalsError,
  } = useQuery({
    queryKey: ['rentals'], // Add filters if needed
    queryFn: fetchAllRentals, // Pass filters if needed
    staleTime: 1000 * 60,
    refetchInterval: 1000 * 60 * 2,
  });

  // --- Sorting Logic using useMemo ---
  const sortedRentals = useMemo(() => {
    let sortableItems = [...rentals];
    if (sortConfig.key !== null) {
      sortableItems.sort((a, b) => {
        const aValue = a[sortConfig.key] ?? '';
        const bValue = b[sortConfig.key] ?? '';

        // Determine comparison type based on the key
        if (sortConfig.key === 'id' || sortConfig.key === 'customer_id' || sortConfig.key === 'car_id') {
          // Numeric comparison
          if (aValue < bValue) return sortConfig.direction === 'ascending' ? -1 : 1;
          if (aValue > bValue) return sortConfig.direction === 'ascending' ? 1 : -1;
          return 0;
        } else if (sortConfig.key === 'pickup_datetime' || sortConfig.key === 'dropoff_datetime' || sortConfig.key === 'booking_date') {
            // Date comparison (handle null booking_date)
            const dateA = aValue ? new Date(aValue) : null;
            const dateB = bValue ? new Date(bValue) : null;

            if (dateA === null && dateB === null) return 0;
            if (dateA === null) return sortConfig.direction === 'ascending' ? -1 : 1; // Nulls first asc, last desc
            if (dateB === null) return sortConfig.direction === 'ascending' ? 1 : -1; // Nulls first asc, last desc

            if (dateA < dateB) return sortConfig.direction === 'ascending' ? -1 : 1;
            if (dateA > dateB) return sortConfig.direction === 'ascending' ? 1 : -1;
            return 0;
        } else {
          // String comparison (case-insensitive) for status, pickup_location
          if (String(aValue).toLowerCase() < String(bValue).toLowerCase()) {
            return sortConfig.direction === 'ascending' ? -1 : 1;
          }
          if (String(aValue).toLowerCase() > String(bValue).toLowerCase()) {
            return sortConfig.direction === 'ascending' ? 1 : -1;
          }
          return 0;
        }
      });
    }
    return sortableItems;
  }, [rentals, sortConfig]);

  // --- Request Sort Function ---
  const requestSort = useCallback((key) => {
    let direction = 'ascending';
    if (sortConfig.key === key && sortConfig.direction === 'ascending') {
      direction = 'descending';
    }
    setSortConfig({ key, direction });
  }, [sortConfig]);

  // Mutation for updating rental status
  const { mutate: changeRentalStatus, isPending: isUpdatingStatus, variables: updatingVariables } = useMutation({
    mutationFn: ({ rentalId, newStatus }) => updateRentalStatus(rentalId, newStatus),
    onSuccess: (data, { rentalId, newStatus }) => {
      setError('');
      queryClient.invalidateQueries({ queryKey: ['rentals'] });
      alert(`Rental ${rentalId} status updated to ${newStatus}!`);
    },
    onError: (err, { rentalId, newStatus }) => {
      setError(`Failed to update status for rental ${rentalId}: ${err.message}`);
    },
  });

  // Handler for status change button clicks
  const handleStatusChange = (rentalId, currentStatus, targetStatus) => {
    let confirmMessage = `Change rental ${rentalId} status from ${currentStatus} to ${targetStatus}?`;
    if (targetStatus === 'Cancelled' || targetStatus === 'Returned') {
        confirmMessage += "\nThis action might make the car available again.";
    }
    if (window.confirm(confirmMessage)) {
      setError('');
      changeRentalStatus({ rentalId, newStatus: targetStatus });
    }
  };

  // Helper functions (getStatusClassName, getActionButtonClassName, getPossibleActions) remain the same
  const getPossibleActions = (status) => { /* ... (same as before) ... */
    switch (status) {
      case 'Booked':    return ['Confirmed', 'Cancelled'];
      case 'Confirmed': return ['Active', 'Cancelled'];
      case 'Active':    return ['Returned', 'Cancelled']; // Staff can cancel active rentals
      default:          return []; // No actions for Returned or Cancelled
    }
  };
  const getStatusClassName = (status) => { /* ... (same as before) ... */
     switch (status) {
      case 'Booked': return 'status-booked';
      case 'Confirmed': return 'status-confirmed';
      case 'Active': return 'status-active';
      case 'Returned': return 'status-returned';
      case 'Cancelled': return 'status-cancelled';
      // Add styles for Pending Verification if needed
      case 'Pending Verification': return 'status-pending-verification'; // Define this class in CSS
      default: return '';
    }
  };
  const getActionButtonClassName = (action) => { /* ... (same as before) ... */
    switch (action) {
      case 'Confirmed': return 'admin-button-success';
      case 'Active': return 'admin-button-info';
      case 'Returned': return 'admin-button-primary';
      case 'Cancelled': return 'admin-button-danger';
      default: return 'admin-button-secondary';
    }
  };

  // Style for sortable header
  const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };

  return (
    <div className="admin-container">
      <div className="admin-header">
        <h2>Rental Management</h2>
        {/* Add filter UI here if needed */}
      </div>

      <ErrorMessage message={error} />
      {isLoadingRentals && <LoadingSpinner />}
      <ErrorMessage message={isErrorRentals ? `Error fetching rentals: ${rentalsError?.message}` : null} />

      {!isLoadingRentals && !isErrorRentals && (
        <div className="admin-table-wrapper">
          <table className="admin-table">
            <thead>
              <tr>
                {/* Sortable Headers */}
                <th style={thSortableStyle} onClick={() => requestSort('id')}>
                  ID <SortIcon direction={sortConfig.key === 'id' ? sortConfig.direction : null} />
                </th>
                <th style={thSortableStyle} onClick={() => requestSort('customer_id')}>
                  Cust. ID <SortIcon direction={sortConfig.key === 'customer_id' ? sortConfig.direction : null} />
                </th>
                <th style={thSortableStyle} onClick={() => requestSort('car_id')}>
                  Car ID <SortIcon direction={sortConfig.key === 'car_id' ? sortConfig.direction : null} />
                </th>
                <th style={thSortableStyle} onClick={() => requestSort('status')}>
                  Status <SortIcon direction={sortConfig.key === 'status' ? sortConfig.direction : null} />
                </th>
                <th style={thSortableStyle} onClick={() => requestSort('pickup_datetime')}>
                  Pickup <SortIcon direction={sortConfig.key === 'pickup_datetime' ? sortConfig.direction : null} />
                </th>
                <th style={thSortableStyle} onClick={() => requestSort('dropoff_datetime')}>
                  Dropoff <SortIcon direction={sortConfig.key === 'dropoff_datetime' ? sortConfig.direction : null} />
                </th>
                <th className="wrap-text">Pickup Loc.</th>
                <th style={thSortableStyle} onClick={() => requestSort('booking_date')}>
                  Booked <SortIcon direction={sortConfig.key === 'booking_date' ? sortConfig.direction : null} />
                </th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {/* Use sortedRentals */}
              {sortedRentals.length === 0 ? (
                <tr className="admin-table-placeholder">
                  <td colSpan="9">No rentals found.</td>
                </tr>
              ) : (
                sortedRentals.map((rental) => {
                  const possibleActions = getPossibleActions(rental.status);
                  const isCurrentlyUpdating = isUpdatingStatus && updatingVariables?.rentalId === rental.id;
                  return (
                    <tr key={rental.id}>
                      <td>{rental.id}</td>
                      <td>{rental.customer_id}</td>
                      <td>{rental.car_id}</td>
                      <td><span className={`status-indicator ${getStatusClassName(rental.status)}`}>{rental.status}</span></td>
                      <td>{new Date(rental.pickup_datetime).toLocaleString()}</td>
                      <td>{new Date(rental.dropoff_datetime).toLocaleString()}</td>
                      <td className="wrap-text">{rental.pickup_location || '-'}</td>
                      <td>{rental.booking_date ? new Date(rental.booking_date).toLocaleDateString() : '-'}</td>
                      <td className="actions admin-action-buttons">
                        {possibleActions.map(action => (
                          <button
                            key={action}
                            onClick={() => handleStatusChange(rental.id, rental.status, action)}
                            className={`admin-button ${getActionButtonClassName(action)} admin-button-sm`}
                            disabled={isCurrentlyUpdating}
                          >
                            {isCurrentlyUpdating && updatingVariables?.newStatus === action ? '...' : action}
                          </button>
                        ))}
                        {possibleActions.length === 0 && <span style={{fontSize: '0.85em', color: 'grey'}}>No actions</span>}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default RentalManagement;
