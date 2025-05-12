import React, { useContext } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import navlogo from '../assets/navlogo.png';
import './Navbar.css';

const Navbar = () => {
  const { isAuthenticated, user, logout, loading } = useContext(AuthContext);
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="navbar">
      <div className="left-links">
        <div className="navlogo">
          <Link to="/">
            <img src={navlogo} alt="Channathat Logo" />
          </Link>
        </div>
        <ul className="nav-links">
          <li><Link to="/">Shop service</Link></li>
          <li><Link to="/promotions">Promotions</Link></li>
          <li><Link to="#">Recommend</Link></li>
          <li><Link to="/contact">Contact</Link></li>
        </ul>
      </div>

      <div className="right-links">
        <ul className="nav-links">
          {loading ? (
            <li>Loading...</li>
          ) : isAuthenticated ? (
            <>
              {user && <li style={{ color: 'white', marginRight: '15px', alignSelf: 'center' }}>Welcome, {user.name || user.email}!</li>}
              <li><Link to="/profile">Profile</Link></li>
              <li><Link to="/rental-history">My Rentals</Link></li>
              <li>
                <button onClick={handleLogout} className="logout-button" style={{ background: 'none', border: 'none', color: 'white', cursor: 'pointer', fontSize: '1rem', padding: '0', marginLeft: '5px' }}>
                  Log Out
                </button>
              </li>
            </>
          ) : (
            <>
              <li><Link to="/login">Log in</Link></li>
              <li><Link to="/signup">Create account</Link></li>
            </>
          )}
        </ul>
      </div>
    </nav>
  );
};

export default Navbar;