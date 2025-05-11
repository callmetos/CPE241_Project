import React from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchAvailableCars, fetchPublicStats } from '../services/apiService'; // เพิ่ม fetchPublicStats
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
import './Home.css';

import redCar from '../assets/redcar.png';
import whiteCar from '../assets/whitecar.png';
import grayCar from '../assets/greycar.png';

const Home = () => {
  const {
    data: featuredCarsData,
    isLoading: isLoadingFeaturedCars,
    isError: isErrorFeaturedCars,
    error: errorFeaturedCars
  } = useQuery({
    queryKey: ['featuredCars'],
    queryFn: () => fetchAvailableCars({ limit: 4 }), // ยังคงดึง 4 คันสำหรับ Featured
    staleTime: 1000 * 60 * 10,
  });

  const featuredCars = featuredCarsData || []; // ถ้า fetchAvailableCars คืน array โดยตรง

  // ดึงข้อมูล Public Stats
  const {
      data: publicStats,
      isLoading: isLoadingPublicStats,
      isError: isErrorPublicStats,
      error: errorPublicStats
  } = useQuery({
      queryKey: ['publicStats'],
      queryFn: fetchPublicStats,
      staleTime: 1000 * 60 * 5, // Cache 5 นาที
  });

  const totalAvailableCars = publicStats?.total_available_cars ?? 0;
  const totalBranches = publicStats?.total_branches ?? 0;

  return (
    <div className="home-container">
      <section className="hero-section">
        <div className="background-home">
          <div className="overlay-content">
            <h1>Welcome to Channathat Rent A Car</h1>
            <p>Your go-to car rental service for all your travel needs!</p>
          </div>
        </div>
      </section>

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

      <section className="home-section featured-cars-section">
          <h2 className="home-section-title">Featured Cars</h2>
          {isLoadingFeaturedCars && <LoadingSpinner />}
          <ErrorMessage message={isErrorFeaturedCars ? `Error loading cars: ${errorFeaturedCars?.message}` : null} />
          {!isLoadingFeaturedCars && !isErrorFeaturedCars && (
              <div className="card-container">
                  {featuredCars.length === 0 ? (
                      <p>No featured cars available right now.</p>
                  ) : (
                      featuredCars.map(car => (
                          <div key={car.id} className="featured-car-card">
                              <img
                                  src={car.image_url || `https://placehold.co/300x180/eee/ccc?text=Car`}
                                  alt={`${car.brand} ${car.model}`}
                                  className="featured-car-img"
                                  onError={(e) => { e.target.onerror = null; e.target.src='https://placehold.co/300x180/f8d7da/721c24?text=No+Image'; }}
                              />
                              <div className="featured-car-info">
                                  <h4>{car.brand} {car.model}</h4>
                                  <p>Price: ฿{car.price_per_day?.toFixed(2)} / day</p>
                              </div>
                              <Link to={`/rental/short-term?car=${car.id}`} className="featured-car-link">
                                  View Details
                              </Link>
                          </div>
                      ))
                  )}
              </div>
          )}
      </section>

      <section className="home-section stats-section">
          <div className="stats-container">
              <div className="stat-item">
                  <span className="stat-number">{isLoadingPublicStats ? '...' : totalAvailableCars}</span>
                  <span className="stat-label">Available Cars</span>
              </div>
              <div className="stat-item">
                  <span className="stat-number">{isLoadingPublicStats ? '...' : totalBranches}</span>
                   <span className="stat-label">Service Locations</span>
              </div>
          </div>
          <ErrorMessage message={isErrorPublicStats ? `Error loading stats: ${errorPublicStats?.message}` : null} />
      </section>

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
            <Link to="/rental/short-term">
                <button className="search-button">Search Cars</button>
            </Link>
        </div>
      </section>
    </div>
  );
};

export default Home;
