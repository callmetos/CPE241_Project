import React, { useState, useEffect } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateCustomer } from '../../services/apiService.js';
import ErrorMessage from '../ErrorMessage.jsx';
import './AdminForm.css'; // Import form styles
import './AdminCommon.css'; // Import common styles

const CustomerForm = ({ initialData, onClose }) => {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({ name: '', email: '', phone: '' });
  const [error, setError] = useState('');

  useEffect(() => {
    if (initialData) {
      setFormData({
        name: initialData.name || '',
        email: initialData.email || '',
        phone: initialData.phone || '',
      });
    }
  }, [initialData]);

  const { mutate: editCustomer, isPending: isUpdating } = useMutation({
    mutationFn: (updatedData) => updateCustomer(initialData.id, updatedData),
    onSuccess: (updatedCustomer) => {
      console.log('Customer updated:', updatedCustomer);
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      alert('Customer updated successfully!');
      onClose();
    },
    onError: (err) => {
      console.error('Error updating customer:', err);
      setError(`Failed to update customer: ${err.message}`);
    },
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
    if (!/\S+@\S+\.\S+/.test(formData.email)) {
      setError('Please enter a valid email address.'); return;
    }
    const payload = {
      name: formData.name.trim(),
      email: formData.email.trim(),
      phone: formData.phone.trim() || null,
    };
    editCustomer(payload);
  };

  const isProcessing = isUpdating;

  if (!initialData) return null;

  return (
    <div className="admin-form-overlay" onClick={onClose}>
      <div className="admin-form-container" onClick={(e) => e.stopPropagation()}>
        <h3 className="admin-form-header">Edit Customer (ID: {initialData.id})</h3>
        <ErrorMessage message={error} />
        <form onSubmit={handleSubmit}>
          <div className="admin-form-group">
            <label htmlFor="name" className="admin-form-label">Name *</label>
            <input
              type="text" id="name" name="name" value={formData.name} onChange={handleChange}
              className="admin-form-input" required disabled={isProcessing}
            />
          </div>
          <div className="admin-form-group">
            <label htmlFor="email" className="admin-form-label">Email *</label>
            <input
              type="email" id="email" name="email" value={formData.email} onChange={handleChange}
              className="admin-form-input" required disabled={isProcessing}
            />
          </div>
          <div className="admin-form-group">
            <label htmlFor="phone" className="admin-form-label">Phone</label>
            <input
              type="tel" id="phone" name="phone" value={formData.phone} onChange={handleChange}
              className="admin-form-input" disabled={isProcessing}
            />
          </div>
          <p className="admin-form-info-text">
            Created: {new Date(initialData.created_at).toLocaleString()} <br />
            Updated: {new Date(initialData.updated_at).toLocaleString()}
          </p>
          <div className="admin-form-button-container">
            <button type="button" onClick={onClose} className="admin-button admin-button-secondary" disabled={isProcessing}>
              Cancel
            </button>
            <button type="submit" className="admin-button admin-button-success" disabled={isProcessing}>
              {isProcessing ? 'Saving...' : 'Update Customer'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CustomerForm;
