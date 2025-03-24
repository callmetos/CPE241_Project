import React, { useState, useEffect } from 'react';
import { getCars } from '../api'; // Named import from api.js

const CarList = () => {
  const [cars, setCars] = useState([]);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchCars = async () => {
      try {
        const response = await getCars(); // Fetch cars from the API
        console.log('Cars fetched:', response.data); // Log the response data
        setCars(response.data); // Set cars in state
      } catch (err) {
        console.error('Error fetching cars:', err); // Log the actual error
        setError('Failed to load cars'); // Set the error state
      }
    };

    fetchCars(); // Call the function to fetch cars when the component mounts
  }, []); // Empty dependency array ensures this runs once when component mounts

  return (
    <div>
      <h2>Available Cars</h2>
      {error && <p style={{ color: 'red' }}>{error}</p>} {/* Display error message */}
      <ul>
        {cars.map((car) => (
          <li key={car.id}>
            {car.brand} {car.model} - ${car.price_per_day} per day
          </li>
        ))}
      </ul>
    </div>
  );
};

export default CarList;
