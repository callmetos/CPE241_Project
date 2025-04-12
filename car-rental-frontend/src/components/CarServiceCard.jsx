import React from 'react';
import { Link } from 'react-router-dom'; // Import Link

// Assuming basic styling here, move to CSS for better management
const cardStyle = {
    width: '30%',
    minWidth: '250px', // Ensure readability on smaller screens
    textAlign: 'center',
    padding: '20px',
    backgroundColor: '#fff',
    boxShadow: '0 4px 8px rgba(0, 0, 0, 0.1)',
    borderRadius: '8px',
    margin: '10px' // Add some margin
};

const imgStyle = {
    width: '100%',
    height: '150px', // Fixed height for consistency
    objectFit: 'cover', // Ensure image covers the area nicely
    borderRadius: '8px',
    marginBottom: '15px'
};

const buttonStyle = {
    backgroundColor: '#007bff',
    color: 'white',
    padding: '10px 20px',
    border: 'none',
    borderRadius: '4px',
    marginTop: '15px',
    cursor: 'pointer',
    textDecoration: 'none' // Remove underline from Link inside button
};


// Example props: title, description, imageUrl, linkTo
const CarServiceCard = ({ title, description, imageUrl, linkTo = "#" }) => {
    return (
        <div style={cardStyle}>
            {imageUrl && <img src={imageUrl} alt={title} style={imgStyle}/>}
            <h3>{title}</h3>
            <p>{description}</p>
            {/* Wrap button content in Link */}
            <Link to={linkTo}>
                <button style={buttonStyle} onMouseOver={(e) => e.target.style.backgroundColor='#0056b3'} onMouseOut={(e) => e.target.style.backgroundColor='#007bff'}>
                    Learn More
                </button>
            </Link>
        </div>
    );
};

export default CarServiceCard;