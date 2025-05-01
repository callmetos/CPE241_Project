import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App'; // Import the main App component

// Import global styles - ensure these are loaded
import './index.css';
// If App.css contains essential layout styles used by App.jsx, import it here too
// import './App.css';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    {/* App component already includes Providers and Router */}
    <App />
  </React.StrictMode>
);
