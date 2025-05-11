import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchCarReviews } from '../services/apiService';
import LoadingSpinner from './LoadingSpinner';
import ErrorMessage from './ErrorMessage';

const StarRating = ({ rating }) => {
    const stars = [];
    for (let i = 1; i <= 5; i++) {
        stars.push(
            <span key={i} style={{ color: i <= rating ? '#ffc107' : '#e0e0e0', fontSize: '1.2em' }}>
                ★
            </span>
        );
    }
    return <div>{stars}</div>;
};

const CarReviews = ({ carId, carBrand, carModel, onClose }) => {
    const {
        data: reviews = [],
        isLoading,
        isError,
        error,
    } = useQuery({
        queryKey: ['carReviews', carId],
        queryFn: () => fetchCarReviews(carId),
        enabled: !!carId, // Query เมื่อ carId มีค่าเท่านั้น
    });

    const modalOverlayStyle = {
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.6)',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        zIndex: 1050,
        padding: '20px',
    };

    const modalContentStyle = {
        backgroundColor: '#fff',
        padding: '25px 30px',
        borderRadius: '8px',
        width: '100%',
        maxWidth: '600px',
        maxHeight: '80vh',
        overflowY: 'auto',
        boxShadow: '0 5px 15px rgba(0,0,0,0.3)',
    };

    const reviewItemStyle = {
        borderBottom: '1px solid #eee',
        padding: '15px 0',
    };

    const reviewHeaderStyle = {
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '8px',
    };

    const commentStyle = {
        marginTop: '8px',
        fontSize: '0.95em',
        color: '#555',
        lineHeight: 1.5,
        whiteSpace: 'pre-wrap', // Preserve line breaks in comments
    };

    const closeButtonStyle = {
        padding: '8px 15px',
        backgroundColor: '#6c757d',
        color: 'white',
        border: 'none',
        borderRadius: '4px',
        cursor: 'pointer',
        marginTop: '20px',
        float: 'right',
    };


    if (!carId) return null;

    return (
        <div style={modalOverlayStyle} onClick={onClose}>
            <div style={modalContentStyle} onClick={(e) => e.stopPropagation()}>
                <h3 style={{ textAlign: 'center', marginBottom: '10px', color: '#333' }}>
                    Reviews for {carBrand} {carModel}
                </h3>
                <p style={{textAlign: 'center', fontSize: '0.9em', color: 'grey', marginBottom: '20px'}}>Car ID: {carId}</p>

                {isLoading && <LoadingSpinner />}
                <ErrorMessage message={isError ? `Error loading reviews: ${error?.message}` : null} />

                {!isLoading && !isError && reviews.length === 0 && (
                    <p style={{ textAlign: 'center', color: '#777', marginTop: '20px' }}>
                        No reviews yet for this car. Be the first to rent and review!
                    </p>
                )}

                {!isLoading && !isError && reviews.length > 0 && (
                    <div>
                        {reviews.map((review) => (
                            <div key={review.id} style={reviewItemStyle}>
                                <div style={reviewHeaderStyle}>
                                    <StarRating rating={review.rating} />
                                    <span style={{ fontSize: '0.8em', color: '#777' }}>
                                        {new Date(review.created_at).toLocaleDateString()}
                                    </span>
                                </div>
                                {review.comment && <p style={commentStyle}>{review.comment}</p>}
                                {/* Optionally display customer name/ID if available and desired */}
                                {/* <p style={{fontSize: '0.8em', color: '#aaa', textAlign: 'right', marginTop: '5px'}}>- Customer {review.customer_id}</p> */}
                            </div>
                        ))}
                    </div>
                )}
                <button style={closeButtonStyle} onClick={onClose}>Close</button>
            </div>
        </div>
    );
};

export default CarReviews;
