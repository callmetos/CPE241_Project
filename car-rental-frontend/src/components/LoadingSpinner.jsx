import React from 'react';

const spinnerStyle = {
  border: '4px solid rgba(0, 0, 0, 0.1)',
  width: '36px',
  height: '36px',
  borderRadius: '50%',
  borderLeftColor: '#09f', // Example color
  animation: 'spin 1s linear infinite',
  margin: '20px auto',
};

const keyframesStyle = `
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;

const LoadingSpinner = () => {
  return (
    <>
      <style>{keyframesStyle}</style>
      <div style={spinnerStyle} aria-label="Loading..."></div>
    </>
  );
};

export default LoadingSpinner;