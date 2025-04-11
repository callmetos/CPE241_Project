import React from 'react';
import { Link } from 'react-router-dom';
import navlogo from '../assets/navlogo.png';
import './Navbar.css'

const Navbar = () => {
  return (
    <nav className="navbar">
    
    <div className="left-links"><div className="navlogo">
        <Link to="/">
            <img src={navlogo} alt="Channathat Logo" />
        </Link>
        </div>
        <ul className="nav-links">
            <li><Link to="/">Shop service</Link></li>
            <li><Link to="#">Promotions</Link></li>
            <li><Link to="#">Recommend</Link></li>
            <li><Link to="#">Contact</Link></li>
        </ul>
    </div>
    <div className="right-links">
        <ul className="nav-links">
            <li><Link to="/login">Log in</Link></li>
            <li><Link to="/signup">Create account</Link></li>
        </ul>
    </div>
    </nav>
  );
}

export default Navbar;
