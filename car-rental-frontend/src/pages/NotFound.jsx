import React from 'react';
import { Link } from 'react-router-dom';

const NotFound = () => {
  const containerStyle={ textAlign: 'center', marginTop: '50px', padding: '20px' };
  const headingStyle={ marginBottom: '15px', color: '#dc3545' };
  const paragraphStyle={ marginBottom: '20px', color: '#6c757d' };
  const linkStyle={ color: '#007bff', textDecoration: 'underline' };
  return ( <div style={containerStyle}> <h2 style={headingStyle}>404 - Page Not Found</h2> <p style={paragraphStyle}>Sorry, the page you are looking for does not exist.</p> <Link to="/" style={linkStyle}>Go back to Home</Link> </div> );
};
export default NotFound;