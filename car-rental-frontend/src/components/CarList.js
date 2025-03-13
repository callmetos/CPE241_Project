import React, { useState, useEffect } from 'react';
import { getCars } from '../api';

const CarList = () => {
  const [cars, setCars] = useState([]);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchCars = async () => {
      try {
        const response = await getCars();
        setCars(response.data);
      } catch (err) {
        setError('Failed to load cars');
      }
    };
    
    fetchCars();
  }, []);

  return (
    <div>
      <h2>Available Cars</h2>
      {error && <p style={{ color: 'red' }}>{error}</p>}
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
