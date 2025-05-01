import React, { useState, useEffect } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createUser, updateUser } from '../../services/apiService.js';
import ErrorMessage from '../ErrorMessage.jsx';
import './AdminForm.css';
import './AdminCommon.css';

const UserForm = ({ initialData, onClose }) => {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    name: '', email: '', password: '', role: 'manager',
  });
  const [error, setError] = useState('');
  const isEditMode = Boolean(initialData);

  useEffect(() => {
    if (isEditMode && initialData) {
      setFormData({
        name: initialData.name || '',
        email: initialData.email || '',
        password: '',
        role: initialData.role || 'manager',
      });
    } else {

      setFormData({ name: '', email: '', password: '', role: 'manager' });
    }
  }, [initialData, isEditMode]);


  const mutationOptions = {
    onSuccess: (data) => {
      const action = isEditMode ? 'updated' : 'created';
      console.log(`User ${action}:`, data);
      queryClient.invalidateQueries({ queryKey: ['users'] });
      alert(`User ${action} successfully!`);
      onClose();
    },
    onError: (err) => {
      const action = isEditMode ? 'updating' : 'creating';
      console.error(`Error ${action} user:`, err);
      setError(`Failed to ${action} user: ${err.message}`);
    },
  };

  const { mutate: addUser, isPending: isCreating } = useMutation({
    mutationFn: createUser, ...mutationOptions,
  });

  const { mutate: editUser, isPending: isUpdating } = useMutation({
    mutationFn: (updatedData) => updateUser(initialData.id, updatedData), ...mutationOptions,
  });


  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    setError('');


    if (!formData.name.trim() || !formData.email.trim()) {
      setError('Name and Email are required.'); return;
    }
    if (!isEditMode && !formData.password) {
      setError('Password is required for new users.'); return;
    }
    if (!isEditMode && formData.password.length < 6) {
      setError('Password must be at least 6 characters long.'); return;
    }
    if (!/\S+@\S+\.\S+/.test(formData.email)) {
      setError('Please enter a valid email address.'); return;
    }
    if (!formData.role) {
        setError('Role is required.'); return;
    }

    let payload = {
      name: formData.name.trim(),
      email: formData.email.trim(),
      role: formData.role,
    };

    if (isEditMode) {

      editUser(payload);
    } else {
      payload.password = formData.password;
      addUser(payload);
    }
  };

  const isProcessing = isCreating || isUpdating;

  return (
    <div className="admin-form-overlay" onClick={onClose}>
      <div className="admin-form-container" onClick={(e) => e.stopPropagation()}>
        <h3 className="admin-form-header">{isEditMode ? 'Edit User' : 'Add New User'}</h3>
        <ErrorMessage message={error} />
        <form onSubmit={handleSubmit}>
          <div className="admin-form-group">
            <label htmlFor="name" className="admin-form-label">Name *</label>
            <input type="text" id="name" name="name" value={formData.name} onChange={handleChange} className="admin-form-input" required disabled={isProcessing} />
          </div>
          <div className="admin-form-group">
            <label htmlFor="email" className="admin-form-label">Email *</label>
            <input type="email" id="email" name="email" value={formData.email} onChange={handleChange} className="admin-form-input" required disabled={isProcessing} />
          </div>
          {!isEditMode && (
            <div className="admin-form-group">
              <label htmlFor="password">Password * (min. 6 chars)</label>
              <input type="password" id="password" name="password" value={formData.password} onChange={handleChange} className="admin-form-input" required={!isEditMode} disabled={isProcessing} />
            </div>
          )}
           <div className="admin-form-group">
            <label htmlFor="role" className="admin-form-label">Role *</label>
            <select id="role" name="role" value={formData.role} onChange={handleChange} className="admin-form-select" required disabled={isProcessing}>
                <option value="manager">Manager</option>
                <option value="admin">Admin</option>
            </select>
          </div>

          {isEditMode && (
             <p className="admin-form-info-text">Password cannot be changed here. Use a password reset flow if needed.</p>
          )}

          <div className="admin-form-button-container">
            <button type="button" onClick={onClose} className="admin-button admin-button-secondary" disabled={isProcessing}>Cancel</button>
            <button type="submit" className="admin-button admin-button-success" disabled={isProcessing}>
              {isProcessing ? 'Saving...' : (isEditMode ? 'Update User' : 'Add User')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default UserForm;