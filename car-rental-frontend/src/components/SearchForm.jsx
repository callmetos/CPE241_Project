import React, { useState } from 'react';
// *** เพิ่ม: Import useQuery และ fetchBranches ***
import { useQuery } from '@tanstack/react-query';
import { fetchBranches } from '../services/apiService.js'; // ตรวจสอบ Path ให้ถูกต้อง
import LoadingSpinner from './LoadingSpinner.jsx'; // Optional: แสดง Loading ตอนดึงข้อมูลสาขา
import ErrorMessage from './ErrorMessage.jsx'; // Optional: แสดง Error ถ้าดึงข้อมูลสาขาไม่ได้

const SearchForm = ({ onSearch, title = "Search Cars" }) => {
  const [pickupLocation, setPickupLocation] = useState('');
  const [returnLocation, setReturnLocation] = useState('');
  const [pickupDate, setPickupDate] = useState('');
  const [returnDate, setReturnDate] = useState('');
  const [error, setError] = useState('');

  // *** เพิ่ม: Fetch Branches data ***
  const {
      data: branches = [],
      isLoading: isLoadingBranches,
      isError: isErrorBranches,
      error: branchesError
  } = useQuery({
      queryKey: ['branchesList'], // Key สำหรับ cache
      queryFn: fetchBranches,
      staleTime: Infinity, // ไม่ต้อง refetch บ่อยๆ ถ้าสาขาไม่เปลี่ยน
      refetchOnWindowFocus: false,
  });

  const handleSubmit = (e) => {
    e.preventDefault();
    setError('');
    if (!pickupLocation) { // ตรวจสอบว่าเลือกจุดรับหรือยัง
      setError('Please select a pickup location.');
      return;
    }
    if (!pickupDate || !returnDate) {
      setError('Please select both pickup and return dates.');
      return;
    }
    if (new Date(returnDate) <= new Date(pickupDate)) {
        setError('Car return date must be after pick up date.');
        return;
    }
     // *** ส่ง pickupLocation และ returnLocation (ซึ่งตอนนี้คือ branch ID) ***
    // ถ้า returnLocation ไม่ได้เลือก ให้ใช้ค่าเดียวกับ pickupLocation
    onSearch({ pickupLocation, returnLocation: returnLocation || pickupLocation, pickupDate, returnDate });
  };

   // Basic Styles (พิจารณาย้ายไป CSS file)
   const formContainerStyle = { width: '100%', maxWidth: '800px', margin: '20px auto', padding: '30px', border: '1px solid #d0d0d0', borderRadius: '15px', backgroundColor: '#f8f9fa', boxShadow: '0 4px 8px rgba(0,0,0,0.1)' };
   const formTitleStyle = { textAlign: 'center', marginBottom: '25px', color: '#333', fontSize: '1.5em' };
   const formRowStyle = { display: 'flex', flexWrap: 'wrap', gap: '20px', marginBottom: '20px' };
   const formGroupStyle = { flex: '1 1 45%', minWidth: '250px' };
   const labelStyle = { display: 'block', fontWeight: '500', marginBottom: '8px', color: '#555' };
   const inputStyle = { width: '100%', padding: '12px', border: '1px solid #ccc', borderRadius: '8px', fontSize: '1rem' };
   const buttonStyle = { display: 'block', width: '200px', margin: '10px auto 0 auto', padding: '12px', backgroundColor: '#28a745', color: 'white', border: 'none', borderRadius: '8px', cursor: 'pointer', fontSize: '1.1rem', fontWeight: 'bold' };
   const errorStyle = { color: 'red', textAlign: 'center', marginBottom: '15px', minHeight: '1.2em' };

  return (
    <div style={formContainerStyle}>
      <h2 style={formTitleStyle}>{title}</h2>
      {/* แสดง Error จากการ Validate หรือการโหลดข้อมูล */}
      <div style={errorStyle}>{error || (isErrorBranches ? `Error loading locations: ${branchesError?.message}` : '')}</div>

      <form onSubmit={handleSubmit}>
        <div style={formRowStyle}>
            {/* Pickup Location Dropdown */}
            <div style={formGroupStyle}>
              <label style={labelStyle} htmlFor="pickupLoc">Pickup location *</label>
              {/* ใช้ข้อมูลจาก branches */}
              <select
                id="pickupLoc"
                style={inputStyle}
                value={pickupLocation}
                onChange={(e) => setPickupLocation(e.target.value)}
                disabled={isLoadingBranches} // Disable ขณะโหลด
                required // ทำให้ต้องเลือก
              >
                  {/* ตัวเลือกเริ่มต้น */}
                  <option value="" disabled>-- Select Location --</option>
                  {isLoadingBranches ? (
                    <option disabled>Loading locations...</option>
                  ) : (
                    // สร้าง option จากข้อมูล branches ที่ได้มา
                    branches.map((branch) => (
                      <option key={branch.id} value={branch.id}>
                          {/* แสดงชื่อสาขา (ซึ่งควรจะเป็น A, B, C, D หลังจากอัปเดต DB) */}
                          {branch.name}
                      </option>
                    ))
                  )}
              </select>
            </div>

            {/* Return Location Dropdown */}
            <div style={formGroupStyle}>
              <label style={labelStyle} htmlFor="returnLoc">Vehicle return location</label>
              {/* ใช้ข้อมูลจาก branches */}
              <select
                id="returnLoc"
                style={inputStyle}
                value={returnLocation}
                onChange={(e) => setReturnLocation(e.target.value)}
                disabled={isLoadingBranches} // Disable ขณะโหลด
              >
                   {/* ตัวเลือกเริ่มต้น (ให้เหมือนที่รับ) */}
                   <option value="">-- Same as pickup --</option>
                   {isLoadingBranches ? (
                     <option disabled>Loading locations...</option>
                   ) : (
                    branches.map((branch) => (
                        <option key={branch.id} value={branch.id}>
                            {branch.name}
                        </option>
                      ))
                   )}
              </select>
            </div>
        </div>
        <div style={formRowStyle}>
            {/* Date Inputs (เหมือนเดิม) */}
            <div style={formGroupStyle}>
              <label style={labelStyle} htmlFor="pickupDate">Car pick up date *</label>
              <input id="pickupDate" type="datetime-local" style={inputStyle} value={pickupDate} onChange={(e) => setPickupDate(e.target.value)} required />
            </div>
            <div style={formGroupStyle}>
              <label style={labelStyle} htmlFor="returnDate">Car return date *</label>
              <input id="returnDate" type="datetime-local" style={inputStyle} value={returnDate} onChange={(e) => setReturnDate(e.target.value)} required />
            </div>
        </div>
        {/* ปุ่ม Search */}
        <button type="submit" style={buttonStyle} onMouseOver={(e) => e.target.style.backgroundColor='#218838'} onMouseOut={(e) => e.target.style.backgroundColor='#28a745'} disabled={isLoadingBranches}>
            Search
        </button>
      </form>
    </div>
  );
};

export default SearchForm;