import React, { useState, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchBranches, deleteBranch } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import BranchForm from '../../components/admin/BranchForm.jsx';
import '../../components/admin/AdminCommon.css';


const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const BranchManagement = () => {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingBranch, setEditingBranch] = useState(null);
  const [error, setError] = useState('');

  const [sortConfig, setSortConfig] = useState({ key: 'id', direction: 'ascending' });


  const {
    data: branches = [],
    isLoading: isLoadingBranches,
    isError: isErrorBranches,
    error: branchesError,
  } = useQuery({
    queryKey: ['branches'],
    queryFn: fetchBranches,
    staleTime: 1000 * 60 * 5,
  });


  const sortedBranches = useMemo(() => {
    let sortableItems = [...branches];
    if (sortConfig.key !== null) {
      sortableItems.sort((a, b) => {
        const aValue = a[sortConfig.key] ?? '';
        const bValue = b[sortConfig.key] ?? '';

        if (sortConfig.key === 'id') {

          if (aValue < bValue) return sortConfig.direction === 'ascending' ? -1 : 1;
          if (aValue > bValue) return sortConfig.direction === 'ascending' ? 1 : -1;
          return 0;
        } else {

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
  }, [branches, sortConfig]);


  const requestSort = useCallback((key) => {
    let direction = 'ascending';
    if (sortConfig.key === key && sortConfig.direction === 'ascending') {
      direction = 'descending';
    }
    setSortConfig({ key, direction });
  }, [sortConfig]);


  const { mutate: removeBranch, isPending: isDeleting } = useMutation({
    mutationFn: deleteBranch,
    onSuccess: (data, branchId) => {
      setError('');
      queryClient.invalidateQueries({ queryKey: ['branches'] });
      alert(`Branch ID ${branchId} deleted successfully!`);
    },
    onError: (err, branchId) => {
      setError(`Failed to delete branch ${branchId}: ${err.message}`);
    },
  });


  const handleAddBranch = () => { setEditingBranch(null); setShowForm(true); setError(''); };
  const handleEditBranch = (branch) => { setEditingBranch(branch); setShowForm(true); setError(''); };
  const handleDeleteBranch = (branchId) => {
    if (window.confirm(`Delete branch ID ${branchId}? This might fail if cars are assigned.`)) {
      setError('');
      removeBranch(branchId);
    }
  };
  const handleCloseForm = () => { setShowForm(false); setEditingBranch(null); };


  const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };

  return (
    <div className="admin-container">
      <div className="admin-header">
        <h2>Branch Management</h2>
        <button onClick={handleAddBranch} className="admin-button admin-button-primary">
          + Add Branch
        </button>
      </div>

      <ErrorMessage message={error} />


      {isErrorBranches && <ErrorMessage message={`Error fetching branches: ${branchesError?.message}`} />}


      <div className="admin-table-wrapper">
        <table className="admin-table">
          <thead>
            <tr>

              <th style={thSortableStyle} onClick={() => requestSort('id')}>
                ID <SortIcon direction={sortConfig.key === 'id' ? sortConfig.direction : null} />
              </th>
              <th style={thSortableStyle} onClick={() => requestSort('name')}>
                Name <SortIcon direction={sortConfig.key === 'name' ? sortConfig.direction : null} />
              </th>
              <th className="wrap-text">Address</th>
              <th>Phone</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoadingBranches ? (
              <tr className="admin-table-placeholder">
                <td colSpan="5">
                  <LoadingSpinner />
                  <p>Loading branches...</p>
                </td>
              </tr>
            ) : sortedBranches.length === 0 && !isErrorBranches ? (
              <tr className="admin-table-placeholder">
                <td colSpan="5">No branches found.</td>
              </tr>
            ) : (
              sortedBranches.map((branch) => (
                <tr key={branch.id}>
                  <td>{branch.id}</td>
                  <td>{branch.name}</td>
                  <td className="wrap-text">{branch.address || '-'}</td>
                  <td>{branch.phone || '-'}</td>
                  <td className="actions admin-action-buttons">
                    <button onClick={() => handleEditBranch(branch)} className="admin-button admin-button-warning admin-button-sm" disabled={isDeleting}>Edit</button>
                    <button
                      onClick={() => handleDeleteBranch(branch.id)}
                      className="admin-button admin-button-danger admin-button-sm"
                      disabled={isDeleting}
                    >

                      {isDeleting ? 'Deleting...' : 'Delete'}
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>



      {showForm && (
        <BranchForm initialData={editingBranch} onClose={handleCloseForm} />
      )}
    </div>
  );
};

export default BranchManagement;