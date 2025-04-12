// src/pages/Profile.jsx
import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchUserProfile } from '../services/apiService';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';

const Profile = () => {
  const {
    data: user,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ['userProfile'],
    queryFn: fetchUserProfile,
    staleTime: 1000 * 60 * 15, // Cache profile for 15 minutes
    retry: 1,
  });

  return (
    <div>
      <h2>User Profile</h2>
      {isLoading && <LoadingSpinner />}
      <ErrorMessage message={isError ? error?.message : null} />

      {user && !isLoading && !isError && (
        <div>
          <p><strong>ID:</strong> {user.id}</p>
          <p><strong>Name:</strong> {user.name}</p>
          <p><strong>Email:</strong> {user.email}</p>
          <p><strong>Phone:</strong> {user.phone || 'Not provided'}</p>
          <p><strong>Joined on:</strong> {new Date(user.created_at).toLocaleDateString()}</p>
          {/* TODO: Add Edit Profile Button/Functionality */}
        </div>
      )}
    </div>
  );
};

export default Profile;