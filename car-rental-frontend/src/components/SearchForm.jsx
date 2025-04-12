import React, { useState } from 'react';

// Added title prop
const SearchForm = ({ onSearch, title = "Search Cars" }) => {
  const [pickupLocation, setPickupLocation] = useState('');
  const [returnLocation, setReturnLocation] = useState('');
  const [pickupDate, setPickupDate] = useState('');
  const [returnDate, setReturnDate] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    setError('');
    if (!pickupDate || !returnDate) {
      setError('Please select both pickup and return dates.');
      return;
    }
    if (new Date(returnDate) <= new Date(pickupDate)) {
        setError('Car return date must be after pick up date.');
        return;
    }
    // Pass all collected data to the parent
    onSearch({ pickupLocation, returnLocation, pickupDate, returnDate });
  };

   // Basic Styles - Consider moving to a CSS file
   const formContainerStyle = { width: '100%', maxWidth: '800px', margin: '20px auto', padding: '30px', border: '1px solid #d0d0d0', borderRadius: '15px', backgroundColor: '#f8f9fa', boxShadow: '0 4px 8px rgba(0,0,0,0.1)' };
   const formTitleStyle = { textAlign: 'center', marginBottom: '25px', color: '#333', fontSize: '1.5em' };
   const formRowStyle = { display: 'flex', flexWrap: 'wrap', gap: '20px', marginBottom: '20px' }; // Added wrap
   const formGroupStyle = { flex: '1 1 45%', minWidth: '250px' }; // Allow wrapping
   const labelStyle = { display: 'block', fontWeight: '500', marginBottom: '8px', color: '#555' };
   const inputStyle = { width: '100%', padding: '12px', border: '1px solid #ccc', borderRadius: '8px', fontSize: '1rem' };
   const buttonStyle = { display: 'block', width: '200px', margin: '10px auto 0 auto', padding: '12px', backgroundColor: '#28a745', color: 'white', border: 'none', borderRadius: '8px', cursor: 'pointer', fontSize: '1.1rem', fontWeight: 'bold' };
   const errorStyle = { color: 'red', textAlign: 'center', marginBottom: '15px', minHeight: '1.2em' }; // Added minHeight

  return (
    <div style={formContainerStyle}>
      <h2 style={formTitleStyle}>{title}</h2>
      {/* Display Error Message */}
      <div style={errorStyle}>{error && error}</div>

      <form onSubmit={handleSubmit}>
        <div style={formRowStyle}>
            <div style={formGroupStyle}>
              <label style={labelStyle} htmlFor="pickupLoc">Pickup location</label>
              <select id="pickupLoc" style={inputStyle} value={pickupLocation} onChange={(e) => setPickupLocation(e.target.value)}>
                  <option value="">-- Select Location --</option>
                  {/* TODO: Populate with actual branches from API */}
                  <option value="Airport">Airport</option>
                  <option value="Downtown Office">Downtown Office</option>
                  <option value="West Branch">West Branch</option>
              </select>
            </div>
            <div style={formGroupStyle}>
              <label style={labelStyle} htmlFor="returnLoc">Vehicle return location</label>
              <select id="returnLoc" style={inputStyle} value={returnLocation} onChange={(e) => setReturnLocation(e.target.value)}>
                   <option value="">-- Same as pickup --</option>
                   {/* TODO: Populate with actual branches */}
                   <option value="Airport">Airport</option>
                   <option value="Downtown Office">Downtown Office</option>
                   <option value="West Branch">West Branch</option>
              </select>
            </div>
        </div>
        <div style={formRowStyle}>
            <div style={formGroupStyle}>
              <label style={labelStyle} htmlFor="pickupDate">Car pick up date</label>
              <input id="pickupDate" type="datetime-local" style={inputStyle} value={pickupDate} onChange={(e) => setPickupDate(e.target.value)} required />
            </div>
            <div style={formGroupStyle}>
              <label style={labelStyle} htmlFor="returnDate">Car return date</label>
              <input id="returnDate" type="datetime-local" style={inputStyle} value={returnDate} onChange={(e) => setReturnDate(e.target.value)} required />
            </div>
        </div>
        <button type="submit" style={buttonStyle} onMouseOver={(e) => e.target.style.backgroundColor='#218838'} onMouseOut={(e) => e.target.style.backgroundColor='#28a745'}>
            Search
        </button>
      </form>
    </div>
  );
};

export default SearchForm;