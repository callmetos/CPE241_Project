import React from 'react';
// Import CSS if you create a separate Footer.css
// import './Footer.css';

const Footer = () => {
    // Basic inline styling, replace with CSS file for better management
    const footerStyle = {
        backgroundColor: '#2d3e50', // Example color from index.css
        color: 'white',
        padding: '20px', // Reduced padding a bit
        textAlign: 'center',
        marginTop: 'auto' // Push footer down if content is short
    };

    const linkStyle = {
        margin: '0 10px',
        textDecoration: 'none',
        color: 'white'
    };

    return (
        <footer style={footerStyle}>
            <div className="footer-content">
                <p>&copy; {new Date().getFullYear()} Channathat Rent A Car | All Rights Reserved</p>
                <p>Contact us: info@channathatrentacar.com</p>
                <div className="social-links" style={{ marginTop: '10px' }}>
                    {/* Add actual links */}
                    <a href="#" style={linkStyle}>Facebook</a>
                    <a href="#" style={linkStyle}>Instagram</a>
                    <a href="#" style={linkStyle}>YouTube</a>
                    <a href="#" style={linkStyle}>TikTok</a>
                </div>
            </div>
        </footer>
    );
};

export default Footer;