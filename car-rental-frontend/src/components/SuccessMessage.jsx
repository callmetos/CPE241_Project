import React from 'react';

const successStyle = {
  color: '#0f5132',
  backgroundColor: '#d1e7dd',
  border: '1px solid #badbcc',
  padding: '10px 15px',
  borderRadius: '4px',
  margin: '15px 0',
  textAlign: 'center',
  fontSize: '0.9em',
};

const SuccessMessage = ({ message }) => {

  if (!message) return null;

  return (
    <div style={successStyle} role="status">
      {message}
    </div>
  );
};

export default SuccessMessage;