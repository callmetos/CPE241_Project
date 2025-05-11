import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAllCustomers, deleteCustomer } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import CustomerForm from '../../components/admin/CustomerForm.jsx';
import '../../components/admin/AdminCommon.css';

const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const CustomerManagement = () => {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingCustomer, setEditingCustomer] = useState(null);
  const [error, setError] = useState('');
  const [sortConfig, setSortConfig] = useState({ key: 'id', direction: 'ASC' });
  const [uiFilters, setUiFilters] = useState({ name: '', email: '', phone: '' });
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(10);

  const activeQueryParams = useMemo(() => ({
    ...uiFilters,
    page: currentPage,
    limit: itemsPerPage,
    sort_by: sortConfig.key,
    sort_dir: sortConfig.direction.toUpperCase(),
  }), [uiFilters, currentPage, itemsPerPage, sortConfig]);

  const {
    data: customersData,
    isLoading: isLoadingCustomers,
    isError: isErrorCustomers,
    error: customersError,
  } = useQuery({
    queryKey: ['customers', activeQueryParams],
    queryFn: () => fetchAllCustomers(activeQueryParams),
    staleTime: 1000 * 60 * 3,
    keepPreviousData: true,
  });

  const currentItemsToDisplay = customersData?.customers || [];
  const totalItems = customersData?.total_count || 0;
  const totalPages = Math.ceil(totalItems / itemsPerPage);

  const requestSort = useCallback((key) => {
    let direction = 'ASC';
    if (sortConfig.key === key && sortConfig.direction === 'ASC') {
      direction = 'DESC';
    }
    setSortConfig({ key, direction });
    setCurrentPage(1);
  }, [sortConfig]);

  const { mutate: removeCustomer, isPending: isDeleting } = useMutation({
    mutationFn: deleteCustomer,
    onSuccess: () => {
      setError('');
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      alert(`Customer deleted successfully!`);
    },
    onError: (err) => {
      setError(`Failed to delete customer: ${err.message}`);
    },
  });

  const handleEditCustomer = (customer) => { setEditingCustomer(customer); setShowForm(true); setError(''); };
  const handleDeleteCustomer = (customerId) => {
    if (window.confirm(`Delete customer ID ${customerId}? This might fail if they have rentals.`)) {
      setError('');
      removeCustomer(customerId);
    }
  };
  const handleCloseForm = () => { setShowForm(false); setEditingCustomer(null); };

  const handleFilterChange = (e) => {
    const { name, value } = e.target;
    setUiFilters(prev => ({ ...prev, [name]: value }));
    setCurrentPage(1);
  };

  const handleClearFilters = () => {
    setUiFilters({ name: '', email: '', phone: '' });
    setCurrentPage(1);
  };

  const paginate = (pageNumber) => {
    if (pageNumber > 0 && pageNumber <= totalPages) {
        setCurrentPage(pageNumber);
    }
  };

  const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };
  const filterSectionStyle = { display: 'flex', flexWrap: 'wrap', gap: '15px', padding: '15px', marginBottom: '20px', backgroundColor: '#f8f9fa', borderRadius: '4px', border: '1px solid var(--admin-border-color)' };
  const filterGroupStyle = { display: 'flex', flexDirection: 'column', flex: '1 1 180px' };
  const filterLabelStyle = { marginBottom: '5px', fontSize: '0.85em', fontWeight: '500', color: 'var(--admin-text-medium)' };
  const filterInputStyle = { padding: '8px', fontSize: '0.9em', border: '1px solid #ccc', borderRadius: '4px' };
  const filterButtonStyle = { alignSelf: 'flex-end', padding: '8px 15px' };
  const paginationContainerStyle = { display: 'flex', justifyContent: 'center', alignItems: 'center', marginTop: '20px', paddingTop: '15px', borderTop: '1px solid #eee' };
  const paginationButtonStyle = (isActive) => ({ margin: '0 5px', padding: '8px 12px', cursor: 'pointer', backgroundColor: isActive ? 'var(--admin-primary)' : '#f0f0f0', color: isActive ? 'white' : '#333', border: `1px solid ${isActive ? 'var(--admin-primary)' : '#ccc'}`, borderRadius: '4px', fontWeight: isActive ? 'bold' : 'normal' });
  const paginationNavButtonStyle = { ...paginationButtonStyle(false), backgroundColor: '#e9ecef' };

  return (
    <div className="admin-container">
      <div className="admin-header">
        <h2>Customer Management</h2>
      </div>
      <div style={filterSectionStyle}>
        <div style={filterGroupStyle}>
            <label htmlFor="filter-name" style={filterLabelStyle}>Filter by Name</label>
            <input type="text" id="filter-name" name="name" style={filterInputStyle} value={uiFilters.name} onChange={handleFilterChange} placeholder="Enter name..."/>
        </div>
        <div style={filterGroupStyle}>
            <label htmlFor="filter-email" style={filterLabelStyle}>Filter by Email</label>
            <input type="text" id="filter-email" name="email" style={filterInputStyle} value={uiFilters.email} onChange={handleFilterChange} placeholder="Enter email..."/>
        </div>
        <div style={filterGroupStyle}>
            <label htmlFor="filter-phone" style={filterLabelStyle}>Filter by Phone</label>
            <input type="text" id="filter-phone" name="phone" style={filterInputStyle} value={uiFilters.phone} onChange={handleFilterChange} placeholder="Enter phone..."/>
        </div>
        <div style={{...filterGroupStyle, flexDirection: 'row', alignItems: 'flex-end', gap: '10px', flexBasis: 'auto' }}>
            <button onClick={handleClearFilters} className="admin-button admin-button-secondary" style={filterButtonStyle} disabled={isLoadingCustomers}>Clear Filters</button>
        </div>
      </div>
      <ErrorMessage message={error || (isErrorCustomers ? `Error: ${customersError?.message}` : null)} />
      {isLoadingCustomers && <LoadingSpinner />}
      {!isLoadingCustomers && (
        <>
        <div className="admin-table-wrapper">
          <table className="admin-table">
            <thead>
              <tr>
                <th style={thSortableStyle} onClick={() => requestSort('id')}>ID <SortIcon direction={sortConfig.key === 'id' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th style={thSortableStyle} onClick={() => requestSort('name')}>Name <SortIcon direction={sortConfig.key === 'name' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th style={thSortableStyle} onClick={() => requestSort('email')}>Email <SortIcon direction={sortConfig.key === 'email' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th>Phone</th>
                <th style={thSortableStyle} onClick={() => requestSort('created_at')}>Joined <SortIcon direction={sortConfig.key === 'created_at' ? sortConfig.direction.toLowerCase() : null} /></th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {currentItemsToDisplay.length === 0 ? (
                <tr className="admin-table-placeholder"><td colSpan="6">No customers found.</td></tr>
              ) : (
                currentItemsToDisplay.map((customer) => (
                  <tr key={customer.id}>
                    <td>{customer.id}</td>
                    <td>{customer.name}</td>
                    <td>{customer.email}</td>
                    <td>{customer.phone || '-'}</td>
                    <td>{new Date(customer.created_at).toLocaleDateString()}</td>
                    <td className="actions admin-action-buttons">
                      <button onClick={() => handleEditCustomer(customer)} className="admin-button admin-button-warning admin-button-sm" disabled={isDeleting}>Edit</button>
                      <button onClick={() => handleDeleteCustomer(customer.id)} className="admin-button admin-button-danger admin-button-sm" disabled={isDeleting}>{isDeleting ? '...' : 'Delete'}</button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {totalPages > 1 && (
            <div style={paginationContainerStyle}>
                <button onClick={() => paginate(currentPage - 1)} disabled={currentPage === 1} style={paginationNavButtonStyle}>&laquo; Previous</button>
                {Array.from({ length: totalPages }, (_, i) => {
                    const pageNumber = i + 1;
                    const showPage = pageNumber === 1 || pageNumber === totalPages || (pageNumber >= currentPage - 2 && pageNumber <= currentPage + 2);
                    const showEllipsisStart = currentPage > 4 && pageNumber === 2 && totalPages > 5;
                    const showEllipsisEnd = currentPage < totalPages - 3 && pageNumber === totalPages -1 && totalPages > 5;
                    if (showEllipsisStart && pageNumber !== 2 && !(pageNumber >= currentPage - 2 && pageNumber <= currentPage + 2)) return <span key={`ellipsis-start-${pageNumber}`} style={{ margin: '0 5px' }}>...</span>;
                    if (showEllipsisEnd && pageNumber !== totalPages -1 && !(pageNumber >= currentPage - 2 && pageNumber <= currentPage + 2)) return <span key={`ellipsis-end-${pageNumber}`} style={{ margin: '0 5px' }}>...</span>;
                    if(showPage) { return (<button key={pageNumber} onClick={() => paginate(pageNumber)} style={paginationButtonStyle(currentPage === pageNumber)}>{pageNumber}</button>); }
                    if (pageNumber === 2 && showEllipsisStart && !(pageNumber >= currentPage - 2 && pageNumber <= currentPage + 2) ) {return <span key={`ellipsis-start-${pageNumber}`} style={{ margin: '0 5px' }}>...</span>;}
                    if (pageNumber === totalPages -1 && showEllipsisEnd && !(pageNumber >= currentPage - 2 && pageNumber <= currentPage + 2) ) {return <span key={`ellipsis-end-${pageNumber}`} style={{ margin: '0 5px' }}>...</span>;}
                    return null;
                })}
                <button onClick={() => paginate(currentPage + 1)} disabled={currentPage === totalPages || totalPages === 0} style={paginationNavButtonStyle}>Next &raquo;</button>
            </div>
        )}
        </>
      )}
      {showForm && editingCustomer && (<CustomerForm initialData={editingCustomer} onClose={handleCloseForm} />)}
    </div>
  );
};
export default CustomerManagement;