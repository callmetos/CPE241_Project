import React, { useState, useEffect, useContext } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchUserProfile, updateUserProfile } from '../services/apiService';
import { AuthContext } from '../context/AuthContext';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
import '../components/admin/AdminForm.css';
import '../components/admin/AdminCommon.css';

const Profile = () => {
  const queryClient = useQueryClient();
  const { user: authUser, login: updateAuthContextUser } = useContext(AuthContext);
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({ name: '', phone: '' });
  const [apiError, setApiError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  const { data: userProfile, isLoading, isError, error: fetchProfileError } = useQuery({
    queryKey: ['userProfile'],
    queryFn: fetchUserProfile,
    staleTime: 1000 * 60 * 15,
    retry: 1,
    enabled: !!authUser,
  });

  useEffect(() => {
    if (userProfile) {
      setFormData({
        name: userProfile.name || '',
        phone: userProfile.phone || '',
      });
    } else if (authUser) {
      setFormData({
        name: authUser.name || '',
        phone: authUser.phone || '',
      });
    }
  }, [userProfile, authUser]);

  const mutation = useMutation({
    mutationFn: updateUserProfile,
    onSuccess: (updatedData) => {
      queryClient.invalidateQueries({ queryKey: ['userProfile'] });
      setSuccessMessage('Profile updated successfully!');
      setApiError('');
      setIsEditing(false);
      if (updateAuthContextUser && updatedData.token) {
           updateAuthContextUser(updatedData.token);
      }
    },
    onError: (error) => {
      setApiError(error.message || 'Failed to update profile.');
      setSuccessMessage('');
    },
  });

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    setApiError('');
    setSuccessMessage('');
    if (!formData.name.trim()) {
      setApiError('Name cannot be empty.');
      return;
    }
    mutation.mutate({ name: formData.name, phone: formData.phone });
  };

  const pageContainerStyle = { maxWidth: '700px', margin: '30px auto', padding: '25px', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 2px 10px rgba(0,0,0,0.08)' };
  const profileHeaderStyle = { textAlign: 'center', marginBottom: '25px', color: '#333', borderBottom: '1px solid #eee', paddingBottom: '15px' };
  const detailItemStyle = { marginBottom: '12px', fontSize: '1rem' };
  const detailLabelStyle = { fontWeight: '600', color: '#555', minWidth: '100px', display: 'inline-block' };

  if (isLoading) return <div style={pageContainerStyle}><LoadingSpinner /></div>;

  return (
    <div style={pageContainerStyle}>
      <h2 style={profileHeaderStyle}>User Profile</h2>
      <ErrorMessage message={apiError || (isError ? fetchProfileError?.message : null)} />
      {successMessage && <div className="admin-button admin-button-success" style={{marginBottom: '15px', textAlign:'center', display:'block'}}>{successMessage}</div>}

      {userProfile && !isEditing && (
        <div>
          <p style={detailItemStyle}><span style={detailLabelStyle}>ID:</span> {userProfile.id}</p>
          <p style={detailItemStyle}><span style={detailLabelStyle}>Name:</span> {userProfile.name}</p>
          <p style={detailItemStyle}><span style={detailLabelStyle}>Email:</span> {userProfile.email}</p>
          <p style={detailItemStyle}><span style={detailLabelStyle}>Phone:</span> {userProfile.phone || 'Not provided'}</p>
          <p style={detailItemStyle}><span style={detailLabelStyle}>Joined:</span> {new Date(userProfile.created_at).toLocaleDateString()}</p>
          <div style={{ marginTop: '25px', textAlign: 'right' }}>
            <button onClick={() => setIsEditing(true)} className="admin-button admin-button-warning">Edit Profile</button>
          </div>
        </div>
      )}

      {isEditing && (
        <form onSubmit={handleSubmit} className="admin-form-container" style={{ boxShadow: 'none', padding: '0', border: 'none' }}>
          <div className="admin-form-group">
            <label htmlFor="name" className="admin-form-label">Name *</label>
            <input
              type="text"
              id="name"
              name="name"
              className="admin-form-input"
              value={formData.name}
              onChange={handleInputChange}
              required
              disabled={mutation.isPending}
            />
          </div>
          <div className="admin-form-group">
            <label htmlFor="email" className="admin-form-label">Email (Cannot be changed)</label>
            <input
              type="email"
              id="email"
              name="email"
              className="admin-form-input"
              value={userProfile?.email || authUser?.email || ''}
              disabled
            />
          </div>
          <div className="admin-form-group">
            <label htmlFor="phone" className="admin-form-label">Phone</label>
            <input
              type="tel"
              id="phone"
              name="phone"
              className="admin-form-input"
              value={formData.phone}
              onChange={handleInputChange}
              disabled={mutation.isPending}
            />
          </div>
          <div className="admin-form-button-container" style={{borderTop: '1px solid #eee', marginTop:'20px', paddingTop: '20px'}}>
            <button type="button" onClick={() => { setIsEditing(false); setApiError(''); setSuccessMessage('');}} className="admin-button admin-button-secondary" disabled={mutation.isPending}>
              Cancel
            </button>
            <button type="submit" className="admin-button admin-button-success" disabled={mutation.isPending}>
              {mutation.isPending ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      )}
      {!userProfile && !isLoading && !isError && <p>Could not load profile information.</p>}
    </div>
  );
};

export default Profile;