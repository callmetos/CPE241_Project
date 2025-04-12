// src/pages/CarRental.jsx
import React, { useState, useContext } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAvailableCars, createRentalBooking } from '../services/apiService';
import { AuthContext } from '../context/AuthContext';

import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
import SearchForm from '../components/SearchForm';

const bannerImages = { /* ... (same as before) ... */ };
const pageTitles = { /* ... (same as before) ... */ };
// Styles (same as before)
const bannerStyle = { /* ... */ };
const carListStyle = { /* ... */ };
const carItemStyle = { /* ... */ };
const imgStyle = { /* ... */ };
const confirmationStyle = { /* ... */ };
const selectButtonStyle = (car, isBooking, searchCriteria) => ({ /* ... */ });


const CarRental = () => {
  const { rentalType = 'default' } = useParams();
  const { isAuthenticated } = useContext(AuthContext);
  const queryClient = useQueryClient();

  const [searchCriteria, setSearchCriteria] = useState({ pickupDate: '', returnDate: '' });
  const [bookingSuccess, setBookingSuccess] = useState(null);

  const currentPageTitle = pageTitles[rentalType] || pageTitles['default'];
  const currentBannerImage = bannerImages[rentalType] || bannerImages['default'];
  const searchFormTitle = pageTitles[rentalType] || "Search Available Cars";

  const {
    data: cars = [],
    isLoading: isLoadingCars,
    isError: isErrorCars,
    error: errorCars,
  } = useQuery({
    queryKey: ['availableCars', rentalType], // Include rentalType in key
    queryFn: () => fetchAvailableCars({ type: rentalType }), // Pass type to fetch function (Backend needs update)
    enabled: !!isAuthenticated,
    staleTime: 1000 * 60 * 5,
  });

  const {
    mutate: bookCar,
    isPending: isBooking,
    isError: isBookingError,
    error: bookingError,
    reset: resetBookingMutation,
  } = useMutation({
    mutationFn: createRentalBooking,
    onSuccess: (data, variables) => {
      console.log('Booking successful:', data);
      setBookingSuccess({
        carId: variables.car_id,
        pickup: variables.pickup_datetime,
        dropoff: variables.dropoff_datetime,
        carDetails: cars.find(c => c.id === variables.car_id)
      });
      queryClient.invalidateQueries({ queryKey: ['availableCars', rentalType] });
      queryClient.invalidateQueries({ queryKey: ['rentalHistory'] });
    },
    onError: (error) => {
      console.error("Booking mutation error:", error);
      setBookingSuccess(null);
    },
  });

  const handleCarSearch = (criteria) => {
      console.log("Search criteria updated:", criteria);
      setBookingSuccess(null);
      resetBookingMutation();
      setSearchCriteria({
          pickupDate: criteria.pickupDate,
          returnDate: criteria.returnDate,
          pickupLocation: criteria.pickupLocation,
          returnLocation: criteria.returnLocation
      });
      // Optional: Refetch if search criteria directly influence the query
      // queryClient.invalidateQueries({ queryKey: ['availableCars', rentalType] });
  };

  const handleSelectCar = (car) => {
    setBookingSuccess(null);
    resetBookingMutation();

    const { pickupDate, returnDate } = searchCriteria;
    if (!pickupDate || !returnDate) {
      alert("Please perform a search with valid dates first."); return;
    }
     const pickupDateTime = new Date(pickupDate);
     const returnDateTime = new Date(returnDate);
     if (returnDateTime <= pickupDateTime) {
        alert("Return date must be after pickup date."); return;
    }

    // Pass ISO strings to mutation
    bookCar({
      car_id: car.id,
      pickup_datetime: pickupDateTime.toISOString(),
      dropoff_datetime: returnDateTime.toISOString(),
      // rental_type: rentalType, // Send type if backend needs it
    });
  };

  return (
    <div>
      <h1 style={{ textAlign: 'center', marginBottom: '20px' }}>{currentPageTitle}</h1>
      <img src={currentBannerImage} alt={`${currentPageTitle} Banner`} style={bannerStyle} />

      <SearchForm onSearch={handleCarSearch} title={searchFormTitle} />

      <ErrorMessage message={isBookingError ? bookingError?.message : null} />
      {bookingSuccess && (
        <div style={confirmationStyle}>
          <h3>Booking Successful!</h3>
          {bookingSuccess.carDetails ? (
             <p>You have booked the {bookingSuccess.carDetails.brand} {bookingSuccess.carDetails.model} from {new Date(bookingSuccess.pickup).toLocaleString()} to {new Date(bookingSuccess.dropoff).toLocaleString()}.</p>
          ) : (
             <p>Booking confirmed for car ID {bookingSuccess.carId} from {new Date(bookingSuccess.pickup).toLocaleString()} to {new Date(bookingSuccess.dropoff).toLocaleString()}.</p>
          )}
          <p>Status: Booked</p>
        </div>
      )}

      <h2 style={{ textAlign: 'center', margin: '30px 0 20px 0' }}>Select Car</h2>
      {isLoadingCars && <LoadingSpinner />}
      <ErrorMessage message={isErrorCars ? errorCars?.message : null} />

      {!isLoadingCars && cars.length === 0 && !isErrorCars && <p style={{ textAlign: 'center' }}>No cars available for {currentPageTitle}.</p>}

      <div style={carListStyle}>
        {!isLoadingCars && cars.length > 0 && cars.map((car) => (
          <div key={car.id} style={carItemStyle}>
            <img
                src={car.image_url || `https://via.placeholder.com/220x150?text=${car.brand}+${car.model}`}
                alt={`${car.brand} ${car.model}`}
                style={imgStyle}
            />
            <h4 style={{ marginBottom: '15px' }}>{car.brand} {car.model}</h4>
            <button
               onClick={() => handleSelectCar(car)}
               disabled={isBooking || !searchCriteria.pickupDate || !searchCriteria.returnDate || !car.availability}
               style={selectButtonStyle(car, isBooking, searchCriteria)}
            >
              {isBooking && !bookingSuccess && !isBookingError ? 'Selecting...' : 'Select'}
            </button>
             {!car.availability && <p style={{color: 'red', fontSize: '0.8em', marginTop: '5px'}}>Booked/Unavailable</p>}
          </div>
        ))}
      </div>
    </div>
  );
};

export default CarRental;