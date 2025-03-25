import React, { useState } from 'react';

const SearchForm = ({ onSearch }) => {
  const [pickupLocation, setPickupLocation] = useState('Airport or Anywhere');
  const [returnLocation, setReturnLocation] = useState('Airport or Anywhere');
  const [pickupDate, setPickupDate] = useState('');
  const [returnDate, setReturnDate] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();

    if (!pickupDate || !returnDate) {
      setError('Please select both pickup and return dates.');
      return;
    }

    // Call onSearch callback passed from parent (Home or CarRental)
    onSearch({ pickupLocation, returnLocation, pickupDate, returnDate });
  };

  return (
    <div className="search-form">
      <h2>Find Your Car</h2>
      {error && <p className="error">{error}</p>}
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Pickup Location</label>
          <select
            value={pickupLocation}
            onChange={(e) => setPickupLocation(e.target.value)}
          >
            <option>Airport or Anywhere</option>
            <option>Location 1</option>
            <option>Location 2</option>
          </select>
        </div>
        <div className="form-group">
          <label>Vehicle Return Location</label>
          <select
            value={returnLocation}
            onChange={(e) => setReturnLocation(e.target.value)}
          >
            <option>Airport or Anywhere</option>
            <option>Location 1</option>
            <option>Location 2</option>
          </select>
        </div>
        <div className="form-group">
          <label>Pick Up Date</label>
          <input
            type="date"
            value={pickupDate}
            onChange={(e) => setPickupDate(e.target.value)}
          />
        </div>
        <div className="form-group">
          <label>Return Date</label>
          <input
            type="date"
            value={returnDate}
            onChange={(e) => setReturnDate(e.target.value)}
          />
        </div>
        <button type="submit">Search</button>
      </form>
    </div>
  );
};

export default SearchForm;
