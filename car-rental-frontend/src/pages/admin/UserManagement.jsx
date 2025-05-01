import React, { useState, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAllUsers, deleteUser } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
// --- FIX: ตรวจสอบว่า Path นี้ถูกต้อง และไฟล์ UserForm.jsx มีอยู่จริง ---
import UserForm from '../../components/admin/UserForm.jsx';
// --- End FIX ---
import '../../components/admin/AdminCommon.css';


const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const UserManagement = () => {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingUser, setEditingUser] = useState(null);
  const [error, setError] = useState('');

  const [sortConfig, setSortConfig] = useState({ key: 'id', direction: 'ascending' });


  const {
    data: users = [],
    isLoading: isLoadingUsers,
    isError: isErrorUsers,
    error: usersError,
  } = useQuery({
    queryKey: ['users'],
    queryFn: fetchAllUsers,
    staleTime: 1000 * 60 * 3,
  });


  const sortedUsers = useMemo(() => {
    let sortableItems = [...users];
    if (sortConfig.key !== null) {
      sortableItems.sort((a, b) => {
        const aValue = a[sortConfig.key] ?? '';
        const bValue = b[sortConfig.key] ?? '';

        if (sortConfig.key === 'id') {

          if (aValue < bValue) return sortConfig.direction === 'ascending' ? -1 : 1;
          if (aValue > bValue) return sortConfig.direction === 'ascending' ? 1 : -1;
          return 0;
        } else if (sortConfig.key === 'created_at') {

            const dateA = new Date(aValue);
            const dateB = new Date(bValue);
            if (dateA < dateB) return sortConfig.direction === 'ascending' ? -1 : 1;
            if (dateA > dateB) return sortConfig.direction === 'ascending' ? 1 : -1;
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
  }, [users, sortConfig]);


  const requestSort = useCallback((key) => {
    let direction = 'ascending';
    if (sortConfig.key === key && sortConfig.direction === 'ascending') {
      direction = 'descending';
    }
    setSortConfig({ key, direction });
  }, [sortConfig]);



  const { mutate: removeUser, isPending: isDeleting } = useMutation({
    mutationFn: deleteUser,
    onSuccess: (data, userId) => {
      setError('');
      queryClient.invalidateQueries({ queryKey: ['users'] });
      alert(`User ID ${userId} deleted successfully!`);
    },
    onError: (err, userId) => {
      setError(`Failed to delete user ${userId}: ${err.message}`);
    },
  });


  const handleAddUser = () => { setEditingUser(null); setShowForm(true); setError(''); };
  const handleEditUser = (user) => { setEditingUser(user); setShowForm(true); setError(''); };
  const handleDeleteUser = (userId) => {
    if (window.confirm(`Are you sure you want to delete user ID ${userId}?`)) {
      setError('');
      removeUser(userId);
    }
  };
  const handleCloseForm = () => { setShowForm(false); setEditingUser(null); };


  const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };

  return (
    <div className="admin-container">
      <div className="admin-header">
        <h2>User Management (Employees)</h2>
        <button onClick={handleAddUser} className="admin-button admin-button-primary">
          + Add User
        </button>
      </div>

      <ErrorMessage message={error} />
      {isLoadingUsers && <LoadingSpinner />}
      <ErrorMessage message={isErrorUsers ? `Error fetching users: ${usersError?.message}` : null} />

      {!isLoadingUsers && !isErrorUsers && (
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
                <th style={thSortableStyle} onClick={() => requestSort('email')}>
                  Email <SortIcon direction={sortConfig.key === 'email' ? sortConfig.direction : null} />
                </th>
                <th style={thSortableStyle} onClick={() => requestSort('role')}>
                  Role <SortIcon direction={sortConfig.key === 'role' ? sortConfig.direction : null} />
                </th>
                <th style={thSortableStyle} onClick={() => requestSort('created_at')}>
                  Created <SortIcon direction={sortConfig.key === 'created_at' ? sortConfig.direction : null} />
                </th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>

              {sortedUsers.length === 0 ? (
                <tr className="admin-table-placeholder">
                  <td colSpan="6">No users found.</td>
                </tr>
              ) : (
                sortedUsers.map((user) => (
                  <tr key={user.id}>
                    <td>{user.id}</td>
                    <td>{user.name}</td>
                    <td>{user.email}</td>
                    <td>{user.role}</td>
                    <td>{new Date(user.created_at).toLocaleDateString()}</td>
                    <td className="actions admin-action-buttons">
                      <button onClick={() => handleEditUser(user)} className="admin-button admin-button-warning admin-button-sm" disabled={isDeleting}>Edit</button>
                      <button onClick={() => handleDeleteUser(user.id)} className="admin-button admin-button-danger admin-button-sm" disabled={isDeleting}>
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


      {showForm && (
        <UserForm initialData={editingUser} onClose={handleCloseForm} />
      )}
    </div>
  );
};

export default UserManagement;