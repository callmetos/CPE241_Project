// src/pages/CarRental.jsx
import React, { useState, useContext, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
// *** แก้ไข: เพิ่ม useMutation และ initiateRentalBooking ***
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAvailableCars, initiateRentalBooking } from '../services/apiService.js'; // *** แก้ไข: เพิ่ม initiateRentalBooking ***
import { AuthContext } from '../context/AuthContext.jsx';

import LoadingSpinner from '../components/LoadingSpinner.jsx';
import ErrorMessage from '../components/ErrorMessage.jsx';
import SearchForm from '../components/SearchForm.jsx';
// import '../components/admin/AdminCommon.css'; // ไม่น่าจะใช้ในหน้านี้

// Default banner/title data (เหมือนเดิม)
const bannerImages = {
    'short-term': '/src/assets/banner_short.jpg', // Adjust path as needed
    'long-term': '/src/assets/banner_long.jpg',   // Adjust path as needed
    'corporate': '/src/assets/banner_corp.jpg',   // Adjust path as needed
    'default': '/src/assets/banner_default.jpg' // Adjust path as needed
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
  const queryClient = useQueryClient(); // *** เพิ่ม QueryClient ***
  const navigate = useNavigate();

  const [searchCriteria, setSearchCriteria] = useState({ pickupDate: '', returnDate: '', pickupLocation: '', returnLocation: '' });
  const [selectedCarId, setSelectedCarId] = useState(null); // *** เพิ่ม State สำหรับติดตามรถที่กำลังจอง ***
  const [initiationError, setInitiationError] = useState(''); // *** เพิ่ม State สำหรับ Error จาก initiateRentalBooking ***

  const currentPageTitle = pageTitles[rentalType] || pageTitles['default'];
  const currentBannerImage = bannerImages[rentalType] || bannerImages['default'];
  const searchFormTitle = `Search ${currentPageTitle}`;

  // --- Fetch Available Cars Query (เหมือนเดิม) ---
  const {
    data: cars = [],
    isLoading: isLoadingCars,
    isError: isErrorCars,
    error: errorCars,
  } = useQuery({
    queryKey: ['availableCars', rentalType, searchCriteria.pickupLocation, searchCriteria.pickupDate, searchCriteria.returnDate],
    queryFn: () => fetchAvailableCars({
        // ส่ง Criteria ไปให้ Backend กรอง (ถ้า Backend รองรับ)
        // type: rentalType, // อาจจะไม่ต้องส่ง type ถ้า filter ด้วยอย่างอื่น
        branch_id: searchCriteria.pickupLocation || undefined, // ส่ง branch_id ถ้าเลือก
        // ส่งวันที่ไปตรวจสอบ Availability ถ้า Backend รองรับ
        // pickup_date: searchCriteria.pickupDate,
        // return_date: searchCriteria.returnDate,
    }),
    enabled: isAuthenticated, // Query เมื่อ Login แล้วเท่านั้น
    staleTime: 1000 * 60 * 5,
  });

   // --- เรียงข้อมูล cars ตาม id (เหมือนเดิม) ---
   const sortedCars = useMemo(() => {
     return [...cars].sort((a, b) => a.id - b.id);
   }, [cars]);

  // --- *** เพิ่ม Mutation สำหรับ Initiate Booking *** ---
  const { mutate: initiateBooking, isPending: isInitiating } = useMutation({
      mutationFn: initiateRentalBooking, // ใช้ function จาก apiService
      onSuccess: (newRental) => {
          console.log('Rental initiation successful:', newRental);
          setInitiationError(''); // เคลียร์ Error เก่า
          if (newRental && newRental.id) {
              // *** นำทางไปยังหน้า Summary โดยใช้ ID จริง ***
              navigate(`/checkout/${newRental.id}/summary`);
          } else {
              // กรณี Backend ไม่คืน ID
              console.error("Rental ID not found in initiation response:", newRental);
              setInitiationError('Failed to get rental details after booking. Please try again.');
          }
      },
      onError: (err) => {
          console.error('Rental initiation failed:', err);
          setInitiationError(`Could not start booking: ${err.message}`);
      },
      onSettled: () => {
          setSelectedCarId(null); // รีเซ็ต ID รถที่กำลังจองเมื่อเสร็จสิ้น (สำเร็จหรือล้มเหลว)
      }
  });


  // --- Event Handlers ---
  const handleCarSearch = (criteria) => {
      console.log("Search criteria updated:", criteria);
      setInitiationError(''); // เคลียร์ Error เมื่อค้นหาใหม่
      setSelectedCarId(null);
      setSearchCriteria({
          pickupDate: criteria.pickupDate,
          returnDate: criteria.returnDate,
          pickupLocation: criteria.pickupLocation,
          returnLocation: criteria.returnLocation || criteria.pickupLocation // ใช้ที่เดียวกับที่รับถ้าไม่ได้เลือก
      });
  };

  // *** แก้ไข handleSelectCar ***
  const handleSelectCar = (car) => {
    setInitiationError(''); // เคลียร์ Error เก่า
    const { pickupDate, returnDate, pickupLocation } = searchCriteria;

    // Validate dates and location from search form (เหมือนเดิม)
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

    // --- เรียก initiateBooking Mutation ---
    setSelectedCarId(car.id); // ตั้ง ID รถที่กำลังจอง เพื่อแสดงสถานะ loading
    initiateBooking({
        car_id: car.id,
        pickup_datetime: pickupDate, // ส่งค่าจาก State ไป
        dropoff_datetime: returnDate,
        pickup_location: pickupLocation // Backend อาจใช้ Branch Address แทนถ้าตรงนี้เป็น null
    });
  };

  // --- Styles (เหมือนเดิม หรือปรับแก้ตามต้องการ) ---
  const pageContainerStyle = { padding: '0 20px 20px 20px', maxWidth: '1400px', margin: '0 auto' };
  const bannerStyle = { width: '100%', height: 'auto', maxHeight: '250px', objectFit: 'cover', marginBottom: '30px', borderRadius: '8px' };
  const carListStyle = { display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '25px', marginTop: '30px' };
  const carItemStyle = { border: '1px solid #e0e0e0', borderRadius: '8px', padding: '15px', backgroundColor: '#fff', boxShadow: '0 2px 5px rgba(0,0,0,0.05)', display: 'flex', flexDirection: 'column' };
  const carInfoStyle = { flexGrow: 1, marginBottom: '15px' };
  const imageContainerStyle = { width: '100%', height: '180px', backgroundColor: '#f0f0f0', borderRadius: '6px', marginBottom: '15px', overflow: 'hidden', display: 'flex', justifyContent: 'center', alignItems: 'center' };
  const imgStyle = { display: 'block', width: '100%', height: '100%', objectFit: 'cover', borderRadius: '6px' };

  // ปรับ Style ปุ่มให้แสดงสถานะ Loading
   const selectButtonStyle = (car, isCurrentlyInitiating) => ({
    padding: '10px 15px',
    border: 'none',
    borderRadius: '4px',
    cursor: (!car.availability || isCurrentlyInitiating) ? 'not-allowed' : 'pointer',
    backgroundColor: isCurrentlyInitiating ? '#ccc' : (car.availability ? '#007bff' : '#6c757d'), // Grey if initiating or unavailable
    color: 'white',
    fontWeight: 'bold',
    transition: 'background-color 0.2s',
    opacity: (!car.availability || isCurrentlyInitiating) ? 0.7 : 1,
    marginTop: 'auto', // Push button to bottom
    ':hover': { // Pseudo-class requires CSS or library like styled-components
      backgroundColor: (!car.availability || isCurrentlyInitiating) ? '#ccc' : '#0056b3',
    }
   });


  return (
    <div style={pageContainerStyle}>
      <h1 style={{ textAlign: 'center', marginBottom: '20px', color: '#333' }}>{currentPageTitle}</h1>
      <img src={currentBannerImage} alt={`${currentPageTitle} Banner`} style={bannerStyle} />
      <SearchForm onSearch={handleCarSearch} title={searchFormTitle} />

      {/* --- *** แสดง Error จาก initiateBooking *** --- */}
      <ErrorMessage message={initiationError} />

      <h2 style={{ textAlign: 'center', margin: '40px 0 20px 0', color: '#444' }}>Select Your Car</h2>
      {isLoadingCars && <LoadingSpinner />}
      <ErrorMessage message={isErrorCars ? `Error loading cars: ${errorCars?.message}` : null} />

      {!isLoadingCars && !isErrorCars && (
        <div style={carListStyle}>
          {sortedCars.length === 0 && <p style={{ textAlign: 'center', gridColumn: '1 / -1' }}>No cars available matching the criteria for {currentPageTitle}.</p>}

          {sortedCars.map((car) => {
            // *** ตรวจสอบว่ารถคันนี้กำลังอยู่ใน process การ initiate หรือไม่ ***
            const isCurrentlyInitiatingThisCar = isInitiating && selectedCarId === car.id;
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
                <button
                  onClick={() => handleSelectCar(car)}
                  // ปรับเงื่อนไข disable ใหม่
                  disabled={!car.availability || !searchCriteria.pickupDate || !searchCriteria.returnDate || !searchCriteria.pickupLocation || isInitiating}
                  // *** ส่งค่า isCurrentlyInitiatingThisCar ไปให้ Style ***
                  style={selectButtonStyle(car, isCurrentlyInitiatingThisCar)}
                >
                  {/* *** เปลี่ยนข้อความปุ่มตามสถานะ *** */}
                  {isCurrentlyInitiatingThisCar ? 'Processing...' : 'Select this Car'}
                </button>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default CarRental;