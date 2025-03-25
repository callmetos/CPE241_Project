import React from 'react';

const CarServiceCard = ({ title, description, imageUrl }) => {
    return (
        <div className="service-card">
            <img src={imageUrl} alt={title} />
            <h3>{title}</h3>
            <p>{description}</p>
            <button>Learn More</button>
        </div>
    );
};

export default CarServiceCard;
