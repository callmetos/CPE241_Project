import React from 'react';
import './Footer.css';
import fbicon from '../assets/fbicon.png' ;
import igicon from '../assets/igicon.png' ;
import yticon from '../assets/yticon.png' ;
import tkicon from '../assets/tkicon.png' ;
import ftlogo from '../assets/ftlogo.png' ;

// Import specific icons from Font Awesome via react-icons

const Footer = () => {
    const footerStyle = {
        backgroundColor: '#2d3e50',
        color: 'white',
        padding: '20px',
        textAlign: 'center',
        marginTop: 'auto',
        width: '100%'
    };

    const linkStyle = {
        margin: '0 10px',
        textDecoration: 'none',
        color: 'white',
        display: 'flex',
        alignItems: 'center'
    };

    return (
        <footer style={footerStyle}>
            <div className="footer-content">
                <div className="footer-text">
                    <img src={ftlogo} style={{ marginRight: '20px' }}/>
                    <p>&copy; {new Date().getFullYear()} Channathat Rent A Car | All Rights Reserved</p>
                    <p style={{ marginLeft: '10px' }}>Contact us: info@channathatrentacar.com</p>
                </div>
                <div className="social-links">
            <a href="https://www.facebook.com/cpe.kmutt" style={linkStyle}><img src={fbicon}/></a>
            <a href="https://www.instagram.com/cpe_studentunion" style={linkStyle}><img src={igicon}/></a>
            <a href="https://www.youtube.com/@CPE-KMUTT" style={linkStyle}><img src={yticon}/></a>
            <a href="https://www.tiktok.com/@comcamp.kmutt" style={linkStyle}><img src={tkicon}/></a>
            </div>
            </div>
        </footer>
    );
};

export default Footer;
