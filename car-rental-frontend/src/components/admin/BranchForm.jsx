import React, { useState, useEffect } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createBranch, updateBranch } from '../../services/apiService.js';
import ErrorMessage from '../ErrorMessage.jsx';
import SuccessMessage from '../SuccessMessage.jsx';
import './AdminForm.css';
import './AdminCommon.css';

const BranchForm = ({ initialData, onClose }) => {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({ name: '', address: '', phone: '' });
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const isEditMode = Boolean(initialData);

  useEffect(() => {
    if (isEditMode && initialData) {
      setFormData({
        name: initialData.name || '',
        address: initialData.address || '',
        phone: initialData.phone || '',
      });
    } else {
      setFormData({ name: '', address: '', phone: '' });
    }

    setError('');
    setSuccess('');
  }, [initialData, isEditMode]);

  const mutationOptions = {
    onSuccess: (data) => {
      const action = isEditMode ? 'updated' : 'created';
      queryClient.invalidateQueries({ queryKey: ['branches'] });
      setSuccess(`Branch ${action} successfully!`);
      setError('');
      if (!isEditMode) {
        setFormData({ name: '', address: '', phone: '' });
      }

      setTimeout(() => {
          onClose();
      }, 1500);
    },
    onError: (err) => {
      const action = isEditMode ? 'updating' : 'creating';
      console.error(`Error ${action} branch:`, err);
      setError(`Failed to ${action} branch: ${err.message}`);
      setSuccess('');
    },
  };

  const { mutate: addBranch, isPending: isCreating } = useMutation({
    mutationFn: createBranch, ...mutationOptions
  });
  const { mutate: editBranch, isPending: isUpdating } = useMutation({
    mutationFn: (updatedData) => updateBranch(initialData.id, updatedData), ...mutationOptions
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    if (!formData.name.trim()) {
      setError('Branch Name is required.');
      return;
    }
    const payload = {
      name: formData.name.trim(),
      address: formData.address.trim() || null,
      phone: formData.phone.trim() || null,
    };
    if (isEditMode) editBranch(payload);
    else addBranch(payload);
  };

  const isProcessing = isCreating || isUpdating;

  return (
    <div className="admin-form-overlay" onClick={onClose}>
      <div className="admin-form-container" onClick={(e) => e.stopPropagation()}>
        <h3 className="admin-form-header">
          {isEditMode ? 'Edit Branch' : 'Add New Branch'}
        </h3>

        <ErrorMessage message={error} />
        <SuccessMessage message={success} />
        <form onSubmit={handleSubmit}>
          <div className="admin-form-group">
            <label htmlFor="name" className="admin-form-label">Branch Name *</label>
            <input
              type="text" id="name" name="name" value={formData.name} onChange={handleChange}
              className="admin-form-input" required disabled={isProcessing}
            />
          </div>
          <div className="admin-form-group">
            <label htmlFor="address" className="admin-form-label">Address</label>
            <textarea
              id="address" name="address" value={formData.address} onChange={handleChange}
              className="admin-form-textarea" disabled={isProcessing}
            />
          </div>
          <div className="admin-form-group">
            <label htmlFor="phone" className="admin-form-label">Phone</label>
            <input
              type="tel" id="phone" name="phone" value={formData.phone} onChange={handleChange}
              className="admin-form-input" disabled={isProcessing}
            />
          </div>
          <div className="admin-form-button-container">
            <button type="button" onClick={onClose} className="admin-button admin-button-secondary" disabled={isProcessing}>
              Cancel
            </button>
            <button type="submit" className="admin-button admin-button-success" disabled={isProcessing}>
              {isProcessing ? 'Saving...' : (isEditMode ? 'Update Branch' : 'Add Branch')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default BranchForm;