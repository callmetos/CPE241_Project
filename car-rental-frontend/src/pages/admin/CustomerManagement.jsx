import React, { useState, useMemo, useCallback } from 'react'; // Import hooks
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAllCustomers, deleteCustomer } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import CustomerForm from '../../components/admin/CustomerForm.jsx';
import '../../components/admin/AdminCommon.css'; // Import common styles

// --- Sort Icon Component (Optional) ---
const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const CustomerManagement = () => {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingCustomer, setEditingCustomer] = useState(null);
  const [error, setError] = useState('');
  // --- State for sorting ---
  const [sortConfig, setSortConfig] = useState({ key: 'id', direction: 'ascending' }); // Default sort

  // Fetch customers data
  const {
    data: customers = [],
    isLoading: isLoadingCustomers,
    isError: isErrorCustomers,
    error: customersError,
  } = useQuery({
    queryKey: ['customers'],
    queryFn: fetchAllCustomers,
    staleTime: 1000 * 60 * 3,
  });

  // --- Sorting Logic using useMemo ---
  const sortedCustomers = useMemo(() => {
    let sortableItems = [...customers];
    if (sortConfig.key !== null) {
      sortableItems.sort((a, b) => {
        const aValue = a[sortConfig.key] ?? '';
        const bValue = b[sortConfig.key] ?? '';

        if (sortConfig.key === 'id') {
          // Numeric comparison
          if (aValue < bValue) return sortConfig.direction === 'ascending' ? -1 : 1;
          if (aValue > bValue) return sortConfig.direction === 'ascending' ? 1 : -1;
          return 0;
        } else if (sortConfig.key === 'created_at') {
            // Date comparison
            const dateA = new Date(aValue);
            const dateB = new Date(bValue);
            if (dateA < dateB) return sortConfig.direction === 'ascending' ? -1 : 1;
            if (dateA > dateB) return sortConfig.direction === 'ascending' ? 1 : -1;
            return 0;
        } else {
          // String comparison (case-insensitive) for name, email
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
  }, [customers, sortConfig]);

  // --- Request Sort Function ---
  const requestSort = useCallback((key) => {
    let direction = 'ascending';
    if (sortConfig.key === key && sortConfig.direction === 'ascending') {
      direction = 'descending';
    }
    setSortConfig({ key, direction });
  }, [sortConfig]);


  // Mutation for deleting a customer
  const { mutate: removeCustomer, isPending: isDeleting } = useMutation({
    mutationFn: deleteCustomer,
    onSuccess: (data, customerId) => {
      setError('');
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      alert(`Customer ID ${customerId} deleted successfully!`);
    },
    onError: (err, customerId) => {
      setError(`Failed to delete customer ${customerId}: ${err.message}`);
    },
  });

  // UI Handlers
  const handleEditCustomer = (customer) => { setEditingCustomer(customer); setShowForm(true); setError(''); };
  const handleDeleteCustomer = (customerId) => {
    if (window.confirm(`Delete customer ID ${customerId}? This might fail if they have active rentals.`)) {
      setError('');
      removeCustomer(customerId);
    }
  };
  const handleCloseForm = () => { setShowForm(false); setEditingCustomer(null); };

  // Style for sortable header
  const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };

  return (
    <div className="admin-container">
      <div className="admin-header">
        <h2>Customer Management</h2>
        {/* Add search input here if needed */}
      </div>

      <ErrorMessage message={error} />
      {isLoadingCustomers && <LoadingSpinner />}
      <ErrorMessage message={isErrorCustomers ? `Error fetching customers: ${customersError?.message}` : null} />

      {!isLoadingCustomers && !isErrorCustomers && (
        <div className="admin-table-wrapper">
          <table className="admin-table">
            <thead>
              <tr>
                {/* Sortable Headers */}
                <th style={thSortableStyle} onClick={() => requestSort('id')}>
                  ID <SortIcon direction={sortConfig.key === 'id' ? sortConfig.direction : null} />
                </th>
                <th style={thSortableStyle} onClick={() => requestSort('name')}>
                  Name <SortIcon direction={sortConfig.key === 'name' ? sortConfig.direction : null} />
                </th>
                <th style={thSortableStyle} onClick={() => requestSort('email')}>
                  Email <SortIcon direction={sortConfig.key === 'email' ? sortConfig.direction : null} />
                </th>
                <th>Phone</th>
                <th style={thSortableStyle} onClick={() => requestSort('created_at')}>
                  Joined <SortIcon direction={sortConfig.key === 'created_at' ? sortConfig.direction : null} />
                </th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {/* Use sortedCustomers */}
              {sortedCustomers.length === 0 ? (
                <tr className="admin-table-placeholder">
                  <td colSpan="6">No customers found.</td>
                </tr>
              ) : (
                sortedCustomers.map((customer) => (
                  <tr key={customer.id}>
                    <td>{customer.id}</td>
                    <td>{customer.name}</td>
                    <td>{customer.email}</td>
                    <td>{customer.phone || '-'}</td>
                    <td>{new Date(customer.created_at).toLocaleDateString()}</td>
                    <td className="actions admin-action-buttons">
                      <button onClick={() => handleEditCustomer(customer)} className="admin-button admin-button-warning admin-button-sm" disabled={isDeleting}>Edit</button>
                      <button onClick={() => handleDeleteCustomer(customer.id)} className="admin-button admin-button-danger admin-button-sm" disabled={isDeleting}>
                        {isDeleting ? '...' : 'Delete'}
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {showForm && editingCustomer && (
        <CustomerForm initialData={editingCustomer} onClose={handleCloseForm} />
      )}
    </div>
  );
};

export default CustomerManagement;
