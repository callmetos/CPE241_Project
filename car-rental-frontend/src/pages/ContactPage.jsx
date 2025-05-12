import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchBranches } from '../services/apiService';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
// Optional: Import icons if you want to use them
// import { FaMapMarkerAlt, FaPhone, FaEnvelope } from 'react-icons/fa';

const ContactPage = () => {
  const {
    data: branches = [], // Default to empty array
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ['allBranchesForContact'], // Unique query key
    queryFn: fetchBranches,
    staleTime: 1000 * 60 * 15, // Cache branch data for 15 minutes
  });

  // Inline styles (consider moving to a CSS file for larger projects)
  const pageContainerStyle = {
    maxWidth: '1000px',
    margin: '30px auto',
    padding: '20px',
    fontFamily: "'Arial', sans-serif",
  };
  const headerStyle = {
    textAlign: 'center',
    color: '#333',
    marginBottom: '40px',
    fontSize: '2.5rem',
    borderBottom: '2px solid #eee',
    paddingBottom: '15px',
  };
  const generalContactStyle = {
    backgroundColor: '#f8f9fa',
    padding: '25px',
    borderRadius: '8px',
    marginBottom: '40px',
    textAlign: 'center',
    boxShadow: '0 2px 8px rgba(0,0,0,0.05)',
  };
  const generalContactHeaderStyle = {
    color: '#007bff',
    marginBottom: '15px',
  }
  const branchGridStyle = {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
    gap: '25px',
  };
  const branchCardStyle = {
    backgroundColor: '#fff',
    border: '1px solid #e0e0e0',
    borderRadius: '8px',
    padding: '20px',
    boxShadow: '0 2px 5px rgba(0,0,0,0.08)',
  };
  const branchNameStyle = {
    color: '#007bff',
    fontSize: '1.4rem',
    marginBottom: '10px',
    borderBottom: '1px solid #f0f0f0',
    paddingBottom: '8px',
  };
  const branchDetailStyle = {
    fontSize: '0.95rem',
    color: '#555',
    marginBottom: '8px',
    lineHeight: '1.6',
    display: 'flex',
    alignItems: 'center',
  };
  const iconStyle = { marginRight: '10px', color: '#007bff', fontSize: '1.1em' };


  return (
    <div style={pageContainerStyle}>
      <h1 style={headerStyle}>Contact Us</h1>

      <div style={generalContactStyle}>
        <h2 style={generalContactHeaderStyle}>Main Office</h2>
        <p style={branchDetailStyle}>
          {/* <FaEnvelope style={iconStyle} /> */}
          Email: info@channathatrentacar.com
        </p>
        <p style={branchDetailStyle}>
          {/* <FaPhone style={iconStyle} /> */}
          Phone: (02) 123-4567
        </p>
        <p style={branchDetailStyle}>
          {/* <FaMapMarkerAlt style={iconStyle} /> */}
          Address: 123 Car Rental Rd, Bangmod, Thung Khru, Bangkok 10140
        </p>
        {/* You can add a Google Maps embed here if desired */}
      </div>

      <h2 style={{ textAlign: 'center', color: '#333', marginBottom: '30px', fontSize: '2rem' }}>Our Branches</h2>

      {isLoading && <LoadingSpinner />}
      <ErrorMessage message={isError ? `Error fetching branches: ${error?.message}` : null} />

      {!isLoading && !isError && branches.length === 0 && (
        <p style={{ textAlign: 'center', color: '#777' }}>No branch information available at the moment.</p>
      )}

      {!isLoading && !isError && branches.length > 0 && (
        <div style={branchGridStyle}>
          {branches.map((branch) => (
            <div key={branch.id} style={branchCardStyle}>
              <h3 style={branchNameStyle}>{branch.name}</h3>
              {branch.address && (
                <p style={branchDetailStyle}>
                  {/* <FaMapMarkerAlt style={iconStyle} /> */}
                  {branch.address}
                </p>
              )}
              {branch.phone && (
                <p style={branchDetailStyle}>
                  {/* <FaPhone style={iconStyle} /> */}
                  {branch.phone}
                </p>
              )}
              {!branch.address && !branch.phone && <p style={branchDetailStyle}>Contact details not available.</p>}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ContactPage;