import React from 'react';
import { Link } from 'react-router-dom';
import './Home.css';

import redCar from '../assets/redcar.png';
import whiteCar from '../assets/whitecar.png';
import grayCar from '../assets/greycar.png';

const Home = () => {
  return (
    <div className="home-container">
      {/* Hero Section */}
      <section className="hero-section">
        <div className="background-home">
          <div className="overlay-content">
            <h1>Welcome to Channathat Rent A Car</h1>
            <p>Your go-to car rental service for all your travel needs!</p>
          </div>
        </div>
      </section>

      {/* Car Rental Options Section */}
      <section className="main-content">
        <div className="service-options">
          <div className="service-option">
            <img src={redCar} alt="Short-term" className="service-img" />
            <Link to="/rental/short-term">
              <button className="rental-button">Short-term car rental</button>
            </Link>
            <p>Car rental for daily/weekly/monthly</p>
          </div>

          <div className="service-option">
            <img src={whiteCar} alt="Long-term" className="service-img" />
            <Link to="/rental/long-term">
              <button className="rental-button">Long-term car rental</button>
            </Link>
            <p>Annual rental (up to 5 years)</p>
          </div>

          <div className="service-option">
            <img src={grayCar} alt="Corporate" className="service-img" />
            <Link to="/rental/corporate">
              <button className="rental-button">Corporate car rental</button>
            </Link>
            <p>Car rental for companies and organizations</p>
          </div>
        </div>
      </section>

      {/* More Content Section */}
      <section className="below-hero-content">
        <div className="services">
          <h2>Our Service</h2>
            <p className="service-description">
                Channathat Rent A Car, We are the car rental service provider for over 1 day.
            </p>

            <h3 className="why-choose">Why choosing Channathat RENT A CAR?</h3>
            <ul className="features-list">
              <li>We provided you more than 1 brand new cars.</li>
              <li>Replacement cars ready in service.</li>
              <li>We standby 24 hours - 7 days to serves all cases.</li>
            </ul>

            <button className="search-button">Search</button>
        </div>
      </section>
    </div>
  );
};

export default Home;
