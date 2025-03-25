import React, { useEffect, useState } from 'react';

const CarRental = () => {
  const [cars, setCars] = useState([]);
  const [error, setError] = useState('');
  const [rentalSuccess, setRentalSuccess] = useState(false);
  const [selectedCar, setSelectedCar] = useState(null);

  // Fetch the available cars from the backend
  useEffect(() => {
    const fetchCars = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/cars', {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('jwt_token')}`, // Send JWT token
          },
        });
        const data = await response.json();
        setCars(data); // Populate available cars
      } catch (err) {
        setError('Error fetching car data');
      }
    };

    fetchCars();
  }, []);

  // Handle the rental booking
  const handleRental = async (car) => {
    setSelectedCar(car); // Set selected car
    const userId = 1; // Replace with actual user ID from context or token

    try {
      const response = await fetch('http://localhost:8080/api/rentals', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('jwt_token')}`, // Include JWT token
        },
        body: JSON.stringify({
          carId: car.id,
          userId: userId,  // Use actual user ID
        }),
      });
      const data = await response.json();

      if (data.success) {
        setRentalSuccess(true); // Show success message
      } else {
        setError('Error renting the car');
      }
    } catch (err) {
      setError('Error renting the car');
    }
  };

  return (
    <div>
      <h2>Available Cars for Rent</h2>
      {error && <p>{error}</p>}
      {rentalSuccess && selectedCar && (
        <div className="confirmation">
          <h3>Rental Successful!</h3>
          <p>You have rented the {selectedCar.brand} {selectedCar.model}.</p>
        </div>
      )}
      <div className="car-list">
        {cars.length > 0 ? (
          cars.map((car) => (
            <div key={car.id} className="car-item">
              <h3>{car.brand} {car.model}</h3>
              <p>${car.price_per_day} per day</p>
              <button onClick={() => handleRental(car)}>Rent Now</button>
            </div>
          ))
        ) : (
          <p>No cars available at the moment</p>
        )}
      </div>
    </div>
  );
};

export default CarRental;
