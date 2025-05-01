import React, { useState } from 'react';
import './admin/AdminForm.css';
import './admin/AdminCommon.css';
import ErrorMessage from './ErrorMessage.jsx';

const ReviewForm = ({ rentalId, carInfo, onSubmit, onClose, isSubmitting }) => {
    const [rating, setRating] = useState(0);
    const [comment, setComment] = useState('');
    const [error, setError] = useState('');
    const [hoverRating, setHoverRating] = useState(0); // For star hover effect

    const handleSubmit = (e) => {
        e.preventDefault();
        setError('');
        if (rating < 1 || rating > 5) {
            setError('Please select a rating between 1 and 5 stars.');
            return;
        }
        // Call the onSubmit prop passed from RentalHistory
        onSubmit({ rentalId, reviewData: { rating, comment } });
    };

    const renderStars = () => {
        return [1, 2, 3, 4, 5].map((star) => (
            <span
                key={star}
                style={{
                    cursor: 'pointer',
                    fontSize: '2rem', // Larger stars
                    color: star <= (hoverRating || rating) ? '#ffc107' : '#e4e5e9', // Gold or grey
                    marginRight: '5px',
                    transition: 'color 0.2s',
                }}
                onClick={() => !isSubmitting && setRating(star)}
                onMouseEnter={() => setHoverRating(star)}
                onMouseLeave={() => setHoverRating(0)}
            >
                ★
            </span>
        ));
    };

    return (
        <div className="admin-form-overlay" onClick={onClose}>
            <div className="admin-form-container" onClick={(e) => e.stopPropagation()}>
                <h3 className="admin-form-header">Submit Review for Rental #{rentalId}</h3>
                {carInfo && <p style={{ textAlign: 'center', marginBottom: '15px', color: '#555' }}>{carInfo}</p>}
                <ErrorMessage message={error} />
                <form onSubmit={handleSubmit}>
                    <div className="admin-form-group">
                        <label className="admin-form-label">Rating *</label>
                        <div style={{ textAlign: 'center', marginBottom: '15px' }}>{renderStars()}</div>
                    </div>
                    <div className="admin-form-group">
                        <label htmlFor="comment" className="admin-form-label">Comment (Optional)</label>
                        <textarea
                            id="comment"
                            name="comment"
                            value={comment}
                            onChange={(e) => setComment(e.target.value)}
                            className="admin-form-textarea"
                            rows="4"
                            disabled={isSubmitting}
                            placeholder="Share your experience..."
                        />
                    </div>
                    <div className="admin-form-button-container">
                        <button type="button" onClick={onClose} className="admin-button admin-button-secondary" disabled={isSubmitting}>
                            Cancel
                        </button>
                        <button type="submit" className="admin-button admin-button-success" disabled={isSubmitting || rating === 0}>
                            {isSubmitting ? 'Submitting...' : 'Submit Review'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default ReviewForm;