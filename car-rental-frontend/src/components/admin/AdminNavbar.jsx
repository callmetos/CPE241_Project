import React, { useContext, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext.jsx';

const navlogo = 'https://placehold.co/50x50/1a2b4f/a9c1ff?text=Adm';

const AdminNavbar = () => {
  const { user, role, logout, loading } = useContext(AuthContext);
  const navigate = useNavigate();
  const [hoveredLink, setHoveredLink] = useState(null);
  const [hoveredLogout, setHoveredLogout] = useState(false);

  const handleLogout = () => {
    logout();
    navigate('/admin/login', { replace: true });
    console.log('Admin logged out.');
  };

  const navStyle = { display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#1a2b4f', padding: '10px 25px', color: 'white', boxShadow: '0 2px 4px rgba(0, 0, 0, 0.3)', position: 'sticky', top: 0, zIndex: 1000 };
  const linkContainerStyle = { display: 'flex', alignItems: 'center' };
  const logoStyle = { width: '45px', height: '45px', marginRight: '15px', display: 'block' };
  const navLinksStyle = { display: 'flex', listStyle: 'none', margin: 0, padding: 0, alignItems: 'center', flexWrap: 'wrap' };
  const navItemStyle = { marginLeft: '18px' };
  const linkStyle = { color: '#e0e0e0', textDecoration: 'none', padding: '8px 10px', borderRadius: '4px', transition: 'background-color 0.2s ease, color 0.2s ease', fontSize: '0.95em', whiteSpace: 'nowrap' };
  const linkHoverStyle = { backgroundColor: 'rgba(255, 255, 255, 0.1)', color: 'white' };
  const userInfoStyle = { marginRight: '15px', color: '#c0c0c0', fontSize: '0.9em', whiteSpace: 'nowrap' };
  const logoutButtonStyle = { background: 'none', border: '1px solid #ff6b6b', color: '#ff6b6b', cursor: 'pointer', fontSize: '0.9rem', padding: '6px 12px', borderRadius: '4px', marginLeft: '10px', transition: 'background-color 0.2s ease, color 0.2s ease' };
  const logoutButtonHoverStyle = { backgroundColor: '#ff6b6b', color: 'white' };

  const getLinkStyle = (key) => ({ ...linkStyle, ...(hoveredLink === key ? linkHoverStyle : {}) });

  return (
    <nav style={navStyle}>
      <div style={linkContainerStyle}>
        <Link to="/admin/dashboard">
          <img src={navlogo} alt="Admin Logo" style={logoStyle} />
        </Link>
        <ul style={navLinksStyle}>
          <li style={navItemStyle}>
            <Link to="/admin/dashboard" style={getLinkStyle('dash')} onMouseEnter={() => setHoveredLink('dash')} onMouseLeave={() => setHoveredLink(null)}>Dashboard</Link>
          </li>
          <li style={navItemStyle}>
            <Link to="/admin/branches" style={getLinkStyle('branch')} onMouseEnter={() => setHoveredLink('branch')} onMouseLeave={() => setHoveredLink(null)}>Branches</Link>
          </li>
          <li style={navItemStyle}>
            <Link to="/admin/cars" style={getLinkStyle('cars')} onMouseEnter={() => setHoveredLink('cars')} onMouseLeave={() => setHoveredLink(null)}>Cars</Link>
          </li>
          <li style={navItemStyle}>
            <Link to="/admin/customers" style={getLinkStyle('cust')} onMouseEnter={() => setHoveredLink('cust')} onMouseLeave={() => setHoveredLink(null)}>Customers</Link>
          </li>
          <li style={navItemStyle}>
            <Link to="/admin/rentals" style={getLinkStyle('rent')} onMouseEnter={() => setHoveredLink('rent')} onMouseLeave={() => setHoveredLink(null)}>Rentals</Link>
          </li>
          <li style={navItemStyle}>
            <Link to="/admin/verify-slips" style={getLinkStyle('verify')} onMouseEnter={() => setHoveredLink('verify')} onMouseLeave={() => setHoveredLink(null)}>Verify Slips</Link>
          </li>
          {role === 'admin' && (
             <li style={navItemStyle}>
                <Link to="/admin/users" style={getLinkStyle('users')} onMouseEnter={() => setHoveredLink('users')} onMouseLeave={() => setHoveredLink(null)}>Users</Link>
            </li>
          )}
        </ul>
      </div>

      <div style={linkContainerStyle}>
        <ul style={{...navLinksStyle, marginLeft: 'auto'}}>
          {loading ? (
            <li style={navItemStyle}><span style={{ color: '#ccc' }}>Loading...</span></li>
          ) : user ? (
            <>
              <li style={userInfoStyle}>
                {user?.name || user?.email} ({role})
              </li>
              <li style={navItemStyle}>
                <button
                    onClick={handleLogout}
                    style={{...logoutButtonStyle, ...(hoveredLogout ? logoutButtonHoverStyle : {})}}
                    onMouseEnter={() => setHoveredLogout(true)}
                    onMouseLeave={() => setHoveredLogout(false)}
                >
                  Log Out
                </button>
              </li>
            </>
          ) : (
            <li style={navItemStyle}>
                <Link to="/admin/login" style={linkStyle}>Login</Link>
            </li>
          )}
        </ul>
      </div>
    </nav>
  );
};

export default AdminNavbar;