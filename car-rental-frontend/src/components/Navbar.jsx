import React from 'react';
import { Link } from 'react-router-dom';

const Navbar = () => {
  return (
    <nav className="navbar">
      <div className="logo">
        <Link to="/">Channathat Rent A Car</Link>
      </div>
      <ul className="nav-links">
        <li><Link to="/">Shop Service</Link></li>
        <li><Link to="#">Promotions</Link></li>
        <li><Link to="#">Recommend</Link></li>
        <li><Link to="#">Contact</Link></li>
        <li><Link to="/login">Log In</Link></li>
        <li><Link to="/signup">Create Account</Link></li>
      </ul>
    </nav>
  );
}

export default Navbar;
