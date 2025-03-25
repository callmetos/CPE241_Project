import React, { useState, useEffect } from 'react';

const RentalHistory = () => {
  const [rentals, setRentals] = useState([]);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchRentalHistory = async () => {
      const token = localStorage.getItem('jwt_token');
      if (!token) return;

      try {
        const response = await fetch('http://localhost:8080/api/rentals/history', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        const data = await response.json();
        setRentals(data);
      } catch (err) {
        setError('Failed to fetch rental history');
      }
    };

    fetchRentalHistory();
  }, []);

  return (
    <div>
      <h2>Rental History</h2>
      {error && <p>{error}</p>}
      {rentals.length > 0 ? (
        <ul>
          {rentals.map((rental) => (
            <li key={rental.id}>
              {rental.carModel} - {rental.startDate} to {rental.endDate} - {rental.status}
            </li>
          ))}
        </ul>
      ) : (
        <p>No rental history available</p>
      )}
    </div>
  );
};

export default RentalHistory;
