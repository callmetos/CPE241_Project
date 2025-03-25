import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
import SignUp from './pages/Signup';
import Login from './pages/Login';
import CarRental from './pages/CarRental';
import PrivateRoute from './components/PrivateRoute';
import Navbar from './components/Navbar';
import Footer from './components/Footer';
import Profile from './pages/Profile'; // Import Profile page
import RentalHistory from './pages/RentalHistory'; // Import Rental History page

const App = () => {
  return (
    <Router>
      <Navbar />
      <div className="main-content">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/signup" element={<SignUp />} />
          <Route path="/login" element={<Login />} />
          <Route path="/car-rental" element={<PrivateRoute><CarRental /></PrivateRoute>} />
          <Route path="/profile" element={<PrivateRoute><Profile /></PrivateRoute>} /> {/* Add Profile route */}
          <Route path="/rental-history" element={<PrivateRoute><RentalHistory /></PrivateRoute>} /> {/* Add Rental History route */}
        </Routes>
      </div>
      <Footer />
    </Router>
  );
};

export default App;
