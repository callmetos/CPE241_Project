import React, { useState, useEffect } from 'react';

const Profile = () => {
  const [user, setUser] = useState(null);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchUserProfile = async () => {
      const token = localStorage.getItem('jwt_token');
      if (!token) return;

      try {
        const response = await fetch('http://localhost:8080/api/profile', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        const data = await response.json();
        setUser(data);
      } catch (err) {
        setError('Failed to fetch user profile');
      }
    };

    fetchUserProfile();
  }, []);

  return (
    <div>
      <h2>User Profile</h2>
      {error && <p>{error}</p>}
      {user ? (
        <div>
          <h3>{user.name}</h3>
          <p>Email: {user.email}</p>
          <p>Joined on: {user.createdAt}</p>
        </div>
      ) : (
        <p>Loading profile...</p>
      )}
    </div>
  );
};

export default Profile;
