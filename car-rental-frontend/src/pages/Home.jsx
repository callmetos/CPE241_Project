import React from 'react';
import SearchForm from '../components/SearchForm';
import { Link } from 'react-router-dom';

const Home = () => {
  // Handle the search action from the SearchForm component
  const handleSearch = (searchData) => {
    console.log('Search criteria:', searchData);
    // You can use the searchData to filter cars or send it to an API endpoint
  };

  return (
    <div className="home-container">
      <h1>Welcome to Channathat Rent A Car</h1>
      <p>Your go-to car rental service for all your travel needs!</p>
      
      {/* Integrate the Search Form */}
      <SearchForm onSearch={handleSearch} />
      
      <div className="service-options">
        <div className="service-option">
          <h3>Short-term car rental</h3>
          <p>Car rental for daily/weekly/monthly</p>
          <button><Link to="/car-rental">Rent Now</Link></button>
        </div>
        <div className="service-option">
          <h3>Long-term car rental</h3>
          <p>Annual rental (5 years)</p>
          <button><Link to="/car-rental">Rent Now</Link></button>
        </div>
        <div className="service-option">
          <h3>Corporate car rental</h3>
          <p>Car rental for companies and organizations</p>
          <button><Link to="/car-rental">Rent Now</Link></button>
        </div>
      </div>
    </div>
  );
};

export default Home;
