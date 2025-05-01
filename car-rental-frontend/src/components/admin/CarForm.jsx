import React, { useState, useEffect } from 'react';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { createCar, updateCar, fetchBranches } from '../../services/apiService.js';
import ErrorMessage from '../ErrorMessage.jsx';
import LoadingSpinner from '../LoadingSpinner.jsx';
import './AdminForm.css'; // Import form styles
import './AdminCommon.css'; // Import common styles

const CarForm = ({ initialData, onClose }) => {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    brand: '', model: '', price_per_day: '', availability: true,
    parking_spot: '', branch_id: '', image_url: '',
  });
  const [error, setError] = useState('');
  const isEditMode = Boolean(initialData);

  const { data: branches = [], isLoading: isLoadingBranches } = useQuery({
    queryKey: ['branches'], queryFn: fetchBranches, staleTime: Infinity,
  });

  useEffect(() => {
    if (isEditMode && initialData) {
      setFormData({
        brand: initialData.brand || '',
        model: initialData.model || '',
        price_per_day: initialData.price_per_day || '',
        availability: initialData.availability === undefined ? true : initialData.availability,
        parking_spot: initialData.parking_spot || '',
        branch_id: initialData.branch_id || '',
        image_url: initialData.image_url || '',
      });
    } else {
      setFormData({ brand: '', model: '', price_per_day: '', availability: true, parking_spot: '', branch_id: '', image_url: '' });
    }
  }, [initialData, isEditMode]);

  const mutationOptions = {
    onSuccess: (data) => {
      const action = isEditMode ? 'updated' : 'created';
      console.log(`Car ${action}:`, data);
      queryClient.invalidateQueries({ queryKey: ['cars'] });
      alert(`Car ${action} successfully!`);
      onClose();
    },
    onError: (err) => {
      const action = isEditMode ? 'updating' : 'creating';
      console.error(`Error ${action} car:`, err);
      setError(`Failed to ${isEditMode ? 'update' : 'create'} car: ${err.message}`);
    },
  };

  const { mutate: addCar, isPending: isCreating } = useMutation({ mutationFn: createCar, ...mutationOptions });
  const { mutate: editCar, isPending: isUpdating } = useMutation({ mutationFn: (d) => updateCar(initialData.id, d), ...mutationOptions });

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({ ...prev, [name]: type === 'checkbox' ? checked : value }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    setError('');
    if (!formData.brand.trim() || !formData.model.trim() || !formData.price_per_day || !formData.branch_id) {
      setError('Brand, Model, Price per Day, and Branch are required.'); return;
    }
    if (isNaN(parseFloat(formData.price_per_day)) || parseFloat(formData.price_per_day) <= 0) {
      setError('Price per Day must be a positive number.'); return;
    }
    if (isNaN(parseInt(formData.branch_id, 10))) {
      setError('Invalid Branch selected.'); return;
    }
    const payload = {
      ...formData,
      price_per_day: parseFloat(formData.price_per_day),
      branch_id: parseInt(formData.branch_id, 10),
      availability: Boolean(formData.availability),
    };
    if (isEditMode) editCar(payload);
    else addCar(payload);
  };

  const isProcessing = isCreating || isUpdating || isLoadingBranches;

  return (
    <div className="admin-form-overlay" onClick={onClose}>
      {/* Use form-lg class for potentially wider form */}
      <div className="admin-form-container form-lg" onClick={(e) => e.stopPropagation()}>
        <h3 className="admin-form-header">{isEditMode ? 'Edit Car' : 'Add New Car'}</h3>
        <ErrorMessage message={error} />
        {isLoadingBranches ? <LoadingSpinner /> : (
          <form onSubmit={handleSubmit}>
            {/* Use grid layout for form fields */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '15px 20px' }}>
              <div className="admin-form-group">
                <label htmlFor="brand" className="admin-form-label">Brand *</label>
                <input type="text" id="brand" name="brand" value={formData.brand} onChange={handleChange} className="admin-form-input" required disabled={isProcessing} />
              </div>
              <div className="admin-form-group">
                <label htmlFor="model" className="admin-form-label">Model *</label>
                <input type="text" id="model" name="model" value={formData.model} onChange={handleChange} className="admin-form-input" required disabled={isProcessing} />
              </div>
              <div className="admin-form-group">
                <label htmlFor="price_per_day" className="admin-form-label">Price/Day *</label>
                <input type="number" step="0.01" id="price_per_day" name="price_per_day" value={formData.price_per_day} onChange={handleChange} className="admin-form-input" required disabled={isProcessing} />
              </div>
              <div className="admin-form-group">
                <label htmlFor="branch_id" className="admin-form-label">Branch *</label>
                <select id="branch_id" name="branch_id" value={formData.branch_id} onChange={handleChange} className="admin-form-select" required disabled={isProcessing}>
                  <option value="">-- Select Branch --</option>
                  {branches.map(branch => (
                    <option key={branch.id} value={branch.id}>{branch.name} (ID: {branch.id})</option>
                  ))}
                </select>
              </div>
              <div className="admin-form-group">
                <label htmlFor="parking_spot" className="admin-form-label">Parking Spot</label>
                <input type="text" id="parking_spot" name="parking_spot" value={formData.parking_spot} onChange={handleChange} className="admin-form-input" disabled={isProcessing} />
              </div>
              <div className="admin-form-group">
                <label htmlFor="image_url" className="admin-form-label">Image URL</label>
                <input type="url" id="image_url" name="image_url" value={formData.image_url} onChange={handleChange} className="admin-form-input" disabled={isProcessing} placeholder="https://..." />
              </div>
            </div>
            <div className="admin-form-checkbox-group">
              <input type="checkbox" id="availability" name="availability" checked={formData.availability} onChange={handleChange} disabled={isProcessing} className="admin-form-checkbox" />
              <label htmlFor="availability" className="admin-form-checkbox-label">Available for Rent</label>
            </div>
            <div className="admin-form-button-container">
              <button type="button" onClick={onClose} className="admin-button admin-button-secondary" disabled={isProcessing}>Cancel</button>
              <button type="submit" className="admin-button admin-button-success" disabled={isProcessing}>
                {isProcessing ? 'Saving...' : (isEditMode ? 'Update Car' : 'Add Car')}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
};

export default CarForm;
