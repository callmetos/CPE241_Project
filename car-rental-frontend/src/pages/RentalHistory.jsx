import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchRentalHistory, submitReview } from '../services/apiService';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
import ReviewForm from '../components/ReviewForm';
import '../components/admin/AdminCommon.css';

const RentalHistory = () => {
  const queryClient = useQueryClient();
  const [showReviewModal, setShowReviewModal] = useState(false);
  const [currentRentalForReview, setCurrentRentalForReview] = useState(null);
  const [reviewError, setReviewError] = useState('');

  const {
    data: rentals = [],
    isLoading,
    isError,
    error: fetchError,
  } = useQuery({
    queryKey: ['rentalHistory'],
    queryFn: fetchRentalHistory,
    staleTime: 1000 * 60 * 2,
  });

  const { mutate: postReview, isPending: isReviewing } = useMutation({
      mutationFn: ({ rentalId, reviewData }) => submitReview(rentalId, reviewData),
      onSuccess: () => {
          alert('Review submitted successfully!');
          setReviewError('');
          setShowReviewModal(false);
          setCurrentRentalForReview(null);
          queryClient.invalidateQueries({ queryKey: ['rentalHistory'] });
      },
      onError: (err) => {
          console.error("Review Submission Error:", err);
          setReviewError(`Failed to submit review: ${err.message}`);
      }
  });

  const handleOpenReviewModal = (rental) => {
      setReviewError('');
      setCurrentRentalForReview(rental);
      setShowReviewModal(true);
  };

  const handleCloseReviewModal = () => {
      setShowReviewModal(false);
      setCurrentRentalForReview(null);
      setReviewError('');
  };

  const handleReviewSubmit = ({ rentalId, reviewData }) => {
     if (!currentRentalForReview || currentRentalForReview.id !== rentalId) return;
     setReviewError('');
     postReview({ rentalId, reviewData });
  };

  const getCarDisplayName = (rental) => {

    if (rental.car && rental.car.brand && rental.car.model) {
        return `${rental.car.brand} ${rental.car.model}`;
    }

    return `Car ID: ${rental.car_id}`;

  };

  const pageContainerStyle = { maxWidth: '1000px', margin: '20px auto', padding: '20px' };
  const listStyle = { listStyle: 'none', padding: 0 };
  const listItemStyle = { border: '1px solid #e0e0e0', marginBottom: '15px', padding: '20px', borderRadius: '8px', backgroundColor: '#fff', boxShadow: '0 2px 4px rgba(0,0,0,0.05)' };
  const detailStyle = { margin: '5px 0', color: '#333', fontSize: '0.95em' };

  // --- FIX: แก้ไขการกำหนด color ---
  const statusStyle = (status) => {
        let backgroundColor;
        let textColor = 'white'; // Default text color

        switch (status) {
            case 'Returned': backgroundColor = '#198754'; break; // Green
            case 'Cancelled': backgroundColor = '#dc3545'; break; // Red
            case 'Confirmed': backgroundColor = '#0dcaf0'; break; // Cyan
            case 'Active': backgroundColor = '#0d6efd'; break; // Blue
            case 'Booked': backgroundColor = '#ffc107'; textColor = '#333'; break; // Yellow, dark text
            case 'Pending': backgroundColor = '#fd7e14'; break; // Orange
            case 'Pending Verification': backgroundColor = '#6f42c1'; break; // Purple
            default: backgroundColor = '#6c757d'; // Grey default
        }

        return {
            fontWeight: 'bold',
            padding: '3px 8px',
            borderRadius: '12px',
            fontSize: '0.85em',
            backgroundColor: backgroundColor,
            color: textColor, // กำหนด color ที่นี่ที่เดียว
        };
   };
   // --- End FIX ---

   const actionButtonStyle = { marginRight: '10px', marginTop: '10px' };

  return (
    <div style={pageContainerStyle}>
      <h2 style={{ textAlign: 'center', marginBottom: '25px', color: '#333' }}>My Rental History</h2>
      {isLoading && <LoadingSpinner />}
      <ErrorMessage message={isError ? `Error fetching history: ${fetchError?.message}` : null} />

      {!isLoading && rentals.length === 0 && !isError && (
        <p style={{ textAlign: 'center' }}>No rental history available.</p>
      )}

      {!isLoading && rentals.length > 0 && (
        <ul style={listStyle}>
          {rentals.map((rental) => (
            <li key={rental.id} style={listItemStyle}>
              <div style={{ display: 'flex', justifyContent: 'space-between', flexWrap: 'wrap', gap: '10px', marginBottom: '10px' }}>
                  <p style={detailStyle}><strong>Rental ID:</strong> {rental.id}</p>
                   <p style={detailStyle}><strong>Status:</strong> <span style={statusStyle(rental.status)}>{rental.status}</span></p>
              </div>
               <p style={detailStyle}><strong>Car:</strong> {getCarDisplayName(rental)}</p>
              <p style={detailStyle}><strong>Pickup:</strong> {new Date(rental.pickup_datetime).toLocaleString()}</p>
              <p style={detailStyle}><strong>Dropoff:</strong> {new Date(rental.dropoff_datetime).toLocaleString()}</p>
              <p style={detailStyle}><strong>Booked/Created:</strong> {rental.booking_date ? new Date(rental.booking_date).toLocaleDateString() : new Date(rental.created_at).toLocaleDateString()}</p>
              <div style={{marginTop: '15px', borderTop: '1px solid #eee', paddingTop: '15px'}}>
                  {rental.status === 'Returned' && (
                    <button
                        onClick={() => handleOpenReviewModal(rental)}
                        className="admin-button admin-button-info admin-button-sm"
                        style={actionButtonStyle}
                        disabled={isReviewing}
                    >
                      Submit Review
                    </button>
                  )}
                   {rental.status !== 'Returned' && <span style={{fontSize: '0.9em', color: 'grey'}}>Review available after return.</span>}
              </div>
            </li>
          ))}
        </ul>
      )}

      {showReviewModal && currentRentalForReview && (
         <ReviewForm
            rentalId={currentRentalForReview.id}
            carInfo={getCarDisplayName(currentRentalForReview)}
            onSubmit={handleReviewSubmit}
            onClose={handleCloseReviewModal}
            isSubmitting={isReviewing}
            initialError={reviewError}
         />
      )}

    </div>
  );
};

export default RentalHistory;