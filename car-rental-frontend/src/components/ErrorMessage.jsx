import React from 'react';

const errorStyle = {
  color: '#721c24', // Darker red text
  backgroundColor: '#f8d7da', // Light pink background
  border: '1px solid #f5c6cb', // Reddish border
  padding: '10px 15px',
  borderRadius: '4px',
  margin: '15px 0',
  textAlign: 'center',
  fontSize: '0.9em',
};

const ErrorMessage = ({ message }) => {
  // Only render if there's a message
  if (!message) return null;

  return (
    <div style={errorStyle} role="alert">
      {message}
    </div>
  );
};

export default ErrorMessage;