import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';

// Import Pages
import Home from './pages/Home';
import SignUp from './pages/SignUp';
import Login from './pages/Login';
import Logout from './pages/Logout';
import CarRental from './pages/CarRental';
import Profile from './pages/Profile';
import RentalHistory from './pages/RentalHistory';

// Import Components
import Navbar from './components/Navbar'; // Navbar แสดงทุกหน้า
import Footer from './components/Footer';
import PrivateRoute from './components/PrivateRoute';

// App component หลัก Render Navbar, Routes, Footer ตรงๆ
const App = () => {
  return (
    <Router>
      <Navbar /> {/* Navbar แสดงนอก Routes */}
      <div className="main-content" style={{ minHeight: 'calc(100vh - 150px)', padding: '20px' }}>
        <Routes>
          {/* Routes */}
          <Route path="/" element={<Home />} />
          <Route path="/signup" element={<SignUp />} />
          <Route path="/login" element={<Login />} />
          <Route path="/logout" element={<Logout />} />
          <Route path="/rental/:rentalType" element={<PrivateRoute><CarRental /></PrivateRoute>} />
          <Route path="/profile" element={<PrivateRoute><Profile /></PrivateRoute>} />
          <Route path="/rental-history" element={<PrivateRoute><RentalHistory /></PrivateRoute>} />
          <Route path="*" element={<div style={{ textAlign: 'center', marginTop: '50px' }}><h2>404 Not Found</h2><p>The page you requested does not exist.</p><Link to="/">Go back to Home</Link></div>} />
        </Routes>
      </div>
      <Footer /> {/* Footer แสดงนอก Routes */}
    </Router>
  );
};

export default App; // Ensure default export