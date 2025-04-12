import React from 'react';
import {Link } from 'react-router-dom';
import './Home.css'; // Link to the CSS file
import redCar from '../assets/redcar.png';
import whiteCar from '../assets/whitecar.png';
import grayCar from '../assets/greycar.png';

const Home = () => {
  return (
    <div className="home-container">
      <div className="background-home">
        <div className="overlay-content">
      <h1>Welcome to Channathat Rent A Car</h1>
      <p>Your go-to car rental service for all your travel needs!</p>

      <div className="service-options">
        <div className="service-option">
        <img src={redCar} alt="Short-term" className="service-img" />
          <div>
          <Link to="/rental/short-term">
            <button className="rental-button">Short-term car rental</button>
          </Link>
            <p>Car rental for daily/weekly/monthly</p>
          </div>
        </div>

        <div className="service-option">
        <img src={whiteCar} alt="Short-term" className="service-img" />
          <div>
          <Link to="/rental/long-term">
            <button className="rental-button">Long-term car rental</button>
          </Link>
            <p>Annual rental (up to 5 years)</p>
          </div>
        </div>

        <div className="service-option">
        <img src={grayCar} alt="Short-term" className="service-img" />
          <div>
          <Link to="/rental/corporate">
            <button className="rental-button">Corporate car rental</button>
          </Link>
            <p>Car rental for companies and organizations</p>
          </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Home;
