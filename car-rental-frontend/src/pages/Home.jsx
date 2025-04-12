import React from 'react';
import { Link } from 'react-router-dom';

const Home = () => {

  // Style for service options container and individual option
  const optionsContainerStyle = {
      display: 'flex',
      justifyContent: 'space-around',
      flexWrap: 'wrap',
      marginTop: '40px',
      gap: '20px'
  };
  const optionStyle = {
      border: '1px solid #ddd',
      padding: '25px',
      borderRadius: '10px',
      textAlign: 'center',
      width: '30%',
      minWidth: '250px',
      backgroundColor: '#f9f9f9',
      boxShadow: '0 2px 5px rgba(0,0,0,0.1)',
      display: 'flex', // Use flexbox for vertical alignment
      flexDirection: 'column', // Stack elements vertically
      justifyContent: 'space-between' // Push button to bottom
  };
   const buttonStyle = {
       backgroundColor: '#007bff',
       color: 'white',
       padding: '10px 20px',
       border: 'none',
       borderRadius: '5px',
       cursor: 'pointer',
       marginTop: '15px',
       fontSize: '1rem'
   };

  return (
    <div className="home-container" style={{ textAlign: 'center' }}>
      <h1>Welcome to Channathat Rent A Car</h1>
      <p>Your go-to car rental service for all your travel needs!</p>

      {/* Service Options with corrected Links */}
      <div className="service-options" style={optionsContainerStyle}>
        <div className="service-option" style={optionStyle}>
            <div> {/* Wrap text content */}
                <h3>Short-term car rental</h3>
                <p>Car rental for daily/weekly/monthly</p>
            </div>
          <Link to="/rental/short-term">
            <button style={buttonStyle} onMouseOver={(e) => e.target.style.backgroundColor='#0056b3'} onMouseOut={(e) => e.target.style.backgroundColor='#007bff'}>
                Rent Now
            </button>
          </Link>
        </div>
        <div className="service-option" style={optionStyle}>
             <div>
                <h3>Long-term car rental</h3>
                <p>Annual rental (up to 5 years)</p>
             </div>
          <Link to="/rental/long-term">
             <button style={buttonStyle} onMouseOver={(e) => e.target.style.backgroundColor='#0056b3'} onMouseOut={(e) => e.target.style.backgroundColor='#007bff'}>
                Rent Now
            </button>
          </Link>
        </div>
        <div className="service-option" style={optionStyle}>
            <div>
                <h3>Corporate car rental</h3>
                <p>Car rental for companies and organizations</p>
            </div>
          <Link to="/rental/corporate">
            <button style={buttonStyle} onMouseOver={(e) => e.target.style.backgroundColor='#0056b3'} onMouseOut={(e) => e.target.style.backgroundColor='#007bff'}>
                Inquire Now {/* Changed button text */}
            </button>
          </Link>
        </div>
      </div>
    </div>
  );
};

export default Home;