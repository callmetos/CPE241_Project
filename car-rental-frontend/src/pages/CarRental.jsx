import React, { useState, useContext, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAvailableCars, initiateRentalBooking } from '../services/apiService.js';
import { AuthContext } from '../context/AuthContext.jsx';

import LoadingSpinner from '../components/LoadingSpinner.jsx';
import ErrorMessage from '../components/ErrorMessage.jsx';
import SearchForm from '../components/SearchForm.jsx';
import CarReviews from '../components/CarReviews.jsx';

const bannerImages = {
    'short-term': '/src/assets/banner_short.jpg',
    'long-term': '/src/assets/banner_long.jpg',
    'corporate': '/src/assets/banner_corp.jpg',
    'default': '/src/assets/banner_default.jpg'
};
const pageTitles = {
    'short-term': 'Short-Term Car Rental',
    'long-term': 'Long-Term Car Rental',
    'corporate': 'Corporate Car Rental',
    'default': 'Car Rental'
};

const CarRental = () => {
  const { rentalType = 'default' } = useParams();
  const { isAuthenticated } = useContext(AuthContext);
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const [searchCriteria, setSearchCriteria] = useState({ pickupDate: '', returnDate: '', pickupLocation: '', returnLocation: '' });
  const [selectedCarIdForBooking, setSelectedCarIdForBooking] = useState(null);
  const [initiationError, setInitiationError] = useState('');

  const [currentPage, setCurrentPage] = useState(1);
  const [carsPerPage, setCarsPerPage] = useState(15);

  const [viewingReviewsForCar, setViewingReviewsForCar] = useState(null);

  const currentPageTitle = pageTitles[rentalType] || pageTitles['default'];
  const currentBannerImage = bannerImages[rentalType] || bannerImages['default'];
  const searchFormTitle = `Search ${currentPageTitle}`;

  const {
    data: carsDataResponse, // เปลี่ยนชื่อเป็น carsDataResponse เพื่อความชัดเจน
    isLoading: isLoadingCars,
    isError: isErrorCars,
    error: errorCars,
  } = useQuery({
    queryKey: ['availableCarsPaginated', rentalType, searchCriteria.pickupLocation, currentPage, carsPerPage],
    queryFn: () => fetchAvailableCars({
        branch_id: searchCriteria.pickupLocation || undefined,
        page: currentPage,
        limit: carsPerPage,
        availability: true,
    }),
    enabled: isAuthenticated,
    staleTime: 1000 * 60 * 2,
    keepPreviousData: true,
  });

   // แก้ไขตรงนี้: ตรวจสอบให้แน่ใจว่า carsDataResponse.cars เป็น array จริงๆ
   const carsToDisplay = Array.isArray(carsDataResponse?.cars) ? carsDataResponse.cars : [];
   const totalCars = Number.isInteger(carsDataResponse?.total_count) ? carsDataResponse.total_count : 0;
   const totalPages = totalCars > 0 && carsPerPage > 0 ? Math.ceil(totalCars / carsPerPage) : 0;


  const { mutate: initiateBooking, isPending: isInitiating } = useMutation({
      mutationFn: initiateRentalBooking,
      onSuccess: (newRental) => {
          setInitiationError('');
          if (newRental && newRental.id) {
              navigate(`/checkout/${newRental.id}/summary`);
          } else {
              setInitiationError('Failed to get rental details after booking. Please try again.');
          }
      },
      onError: (err) => {
          setInitiationError(`Could not start booking: ${err.message}`);
      },
      onSettled: () => {
          setSelectedCarIdForBooking(null);
      }
  });


  const handleCarSearch = (criteria) => {
      setInitiationError('');
      setSelectedCarIdForBooking(null);
      setViewingReviewsForCar(null);
      setCurrentPage(1);
      setSearchCriteria({
          pickupDate: criteria.pickupDate,
          returnDate: criteria.returnDate,
          pickupLocation: criteria.pickupLocation,
          returnLocation: criteria.returnLocation || criteria.pickupLocation
      });
  };

  const handleSelectCar = (car) => {
    setInitiationError('');
    const { pickupDate, returnDate, pickupLocation } = searchCriteria;
    if (!pickupDate || !returnDate) {
      alert("Please perform a search with valid pickup and return dates first."); return;
    }
     if (!pickupLocation) {
      alert("Please select a pickup location."); return;
    }
     const pickupDateTime = new Date(pickupDate);
     const returnDateTime = new Date(returnDate);
     if (returnDateTime <= pickupDateTime) {
        alert("Return date must be after pickup date."); return;
    }
    setSelectedCarIdForBooking(car.id);
    initiateBooking({
        car_id: car.id,
        pickup_datetime: pickupDate,
        dropoff_datetime: returnDate,
        pickup_location: pickupLocation
    });
  };

  const handleViewReviews = (car) => {
    setViewingReviewsForCar({id: car.id, brand: car.brand, model: car.model });
  };
  const handleCloseReviews = () => {
    setViewingReviewsForCar(null);
  };

  const paginate = (pageNumber) => {
    if (pageNumber > 0 && pageNumber <= totalPages) {
        setCurrentPage(pageNumber);
        window.scrollTo(0, 0);
    }
  };

  const pageContainerStyle = { padding: '0 20px 20px 20px', maxWidth: '1400px', margin: '0 auto' };
  const bannerStyle = { width: '100%', height: 'auto', maxHeight: '250px', objectFit: 'cover', marginBottom: '30px', borderRadius: '8px' };
  const carListStyle = { display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '25px', marginTop: '30px' };
  const carItemStyle = { border: '1px solid #e0e0e0', borderRadius: '8px', padding: '15px', backgroundColor: '#fff', boxShadow: '0 2px 5px rgba(0,0,0,0.05)', display: 'flex', flexDirection: 'column' };
  const carInfoStyle = { flexGrow: 1, marginBottom: '10px' };
  const imageContainerStyle = { width: '100%', height: '180px', backgroundColor: '#f0f0f0', borderRadius: '6px', marginBottom: '15px', overflow: 'hidden', display: 'flex', justifyContent: 'center', alignItems: 'center' };
  const imgStyle = { display: 'block', width: '100%', height: '100%', objectFit: 'cover', borderRadius: '6px' };
  const carActionsStyle = { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 'auto' };

   const selectButtonStyle = (car, isCurrentlyInitiating) => ({
    padding: '8px 12px',
    border: 'none',
    borderRadius: '4px',
    cursor: (!car.availability || isCurrentlyInitiating) ? 'not-allowed' : 'pointer',
    backgroundColor: isCurrentlyInitiating ? '#ccc' : (car.availability ? '#007bff' : '#6c757d'),
    color: 'white',
    fontWeight: '500',
    fontSize: '0.9em',
    transition: 'background-color 0.2s',
    opacity: (!car.availability || isCurrentlyInitiating) ? 0.7 : 1,
   });
   const viewReviewsButtonStyle = {
    padding: '8px 12px',
    border: '1px solid #007bff',
    backgroundColor: 'transparent',
    color: '#007bff',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '0.9em',
    fontWeight: '500',
    transition: 'background-color 0.2s, color 0.2s',
   };


   const paginationContainerStyle = { display: 'flex', justifyContent: 'center', alignItems: 'center', marginTop: '30px', paddingTop: '15px', borderTop: '1px solid #eee' };
   const paginationButtonStyle = (isActive) => ({ margin: '0 5px', padding: '8px 12px', cursor: 'pointer', backgroundColor: isActive ? '#007bff' : '#f0f0f0', color: isActive ? 'white' : '#333', border: `1px solid ${isActive ? '#007bff' : '#ccc'}`, borderRadius: '4px', fontWeight: isActive ? 'bold' : 'normal', });
   const paginationNavButtonStyle = { ...paginationButtonStyle(false), backgroundColor: '#e9ecef' }

  return (
    <div style={pageContainerStyle}>
      <h1 style={{ textAlign: 'center', marginBottom: '20px', color: '#333' }}>{currentPageTitle}</h1>
      <img src={currentBannerImage} alt={`${currentPageTitle} Banner`} style={bannerStyle} />
      <SearchForm onSearch={handleCarSearch} title={searchFormTitle} />
      <ErrorMessage message={initiationError} />

      <h2 style={{ textAlign: 'center', margin: '40px 0 20px 0', color: '#444' }}>Select Your Car</h2>
      {isLoadingCars && <LoadingSpinner />}
      <ErrorMessage message={isErrorCars ? `Error loading cars: ${errorCars?.message}` : null} />

      {!isLoadingCars && !isErrorCars && (
        <>
          <div style={carListStyle}>
            {carsToDisplay.length === 0 && <p style={{ textAlign: 'center', gridColumn: '1 / -1' }}>No cars available matching the criteria for {currentPageTitle}.</p>}
            {carsToDisplay.map((car) => {
              const isCurrentlyInitiatingThisCar = isInitiating && selectedCarIdForBooking === car.id;
              return (
                <div key={car.id} style={carItemStyle}>
                  <div style={imageContainerStyle}>
                    <img
                      src={car.image_url || `https://placehold.co/300x180/f0f0f0/ccc?text=Car+Image`}
                      alt={`${car.brand} ${car.model}`}
                      style={imgStyle}
                      loading="lazy"
                      onError={(e) => { e.target.onerror = null; e.target.src='https://placehold.co/300x180/f8d7da/721c24?text=Image+Error'; }}
                    />
                  </div>
                  <div style={carInfoStyle}>
                    <h4 style={{ marginBottom: '8px', color: '#333' }}>{car.brand} {car.model}</h4>
                    <p style={{ fontSize: '0.9em', color: '#666', marginBottom: '5px' }}>Price: ฿{car.price_per_day?.toFixed(2)} / day</p>
                    {!car.availability && <p style={{color: 'red', fontSize: '0.9em', fontWeight: 'bold', marginTop: '8px'}}>Currently Unavailable</p>}
                  </div>
                  <div style={carActionsStyle}>
                    <button
                        onClick={() => handleViewReviews(car)}
                        style={viewReviewsButtonStyle}
                        onMouseOver={(e) => { e.target.style.backgroundColor = '#007bff'; e.target.style.color = 'white';}}
                        onMouseOut={(e) => { e.target.style.backgroundColor = 'transparent'; e.target.style.color = '#007bff';}}
                    >
                        View Reviews
                    </button>
                    <button
                      onClick={() => handleSelectCar(car)}
                      disabled={!car.availability || !searchCriteria.pickupDate || !searchCriteria.returnDate || !searchCriteria.pickupLocation || isInitiating}
                      style={selectButtonStyle(car, isCurrentlyInitiatingThisCar)}
                    >
                      {isCurrentlyInitiatingThisCar ? 'Processing...' : 'Select Car'}
                    </button>
                  </div>
                </div>
              );
            })}
          </div>

          {totalPages > 0 && ( // Show pagination only if there are pages
            <div style={paginationContainerStyle}>
                <button onClick={() => paginate(currentPage - 1)} disabled={currentPage === 1} style={paginationNavButtonStyle}>&laquo; Previous</button>
                {Array.from({ length: totalPages }, (_, i) => {
                    const pageNumber = i + 1;
                    const pageRange = 2;
                    const showPage = pageNumber === 1 || pageNumber === totalPages || (pageNumber >= currentPage - pageRange && pageNumber <= currentPage + pageRange);
                    if (showPage) { return (<button key={pageNumber} onClick={() => paginate(pageNumber)} style={paginationButtonStyle(currentPage === pageNumber)}>{pageNumber}</button>); }
                    else if (pageNumber === currentPage - pageRange - 1 || pageNumber === currentPage + pageRange + 1) {
                        if ((pageNumber === 2 && currentPage > 4 && totalPages > 5) || (pageNumber === totalPages - 1 && currentPage < totalPages - 3 && totalPages > 5)) {
                            return <span key={`ellipsis-${pageNumber}`} style={{ margin: '0 5px' }}>...</span>;
                        }
                    }
                    return null;
                })}
                <button onClick={() => paginate(currentPage + 1)} disabled={currentPage === totalPages || totalPages === 0} style={paginationNavButtonStyle}>Next &raquo;</button>
            </div>
          )}
        </>
      )}
      {viewingReviewsForCar && (
        <CarReviews
            carId={viewingReviewsForCar.id}
            carBrand={viewingReviewsForCar.brand}
            carModel={viewingReviewsForCar.model}
            onClose={handleCloseReviews}
        />
      )}
    </div>
  );
};

export default CarRental;
