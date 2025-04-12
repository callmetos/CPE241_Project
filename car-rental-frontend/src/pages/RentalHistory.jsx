// src/pages/RentalHistory.jsx
import React from 'react';
import { useQuery /*, useMutation, useQueryClient */ } from '@tanstack/react-query'; // Ready for mutations
import { fetchRentalHistory /*, cancelRental, submitReview */ } from '../services/apiService'; // Ready for mutations
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';

const RentalHistory = () => {
  // const queryClient = useQueryClient(); // Needed for mutations

  const {
    data: rentals = [],
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ['rentalHistory'],
    queryFn: fetchRentalHistory,
    staleTime: 1000 * 60 * 2,
  });

  // --- Example Mutation Hooks (Implement fully if needed) ---
  /*
  const { mutate: cancelBooking, isPending: isCancelling } = useMutation({
      mutationFn: cancelRental, // Assumes cancelRental(rentalId) exists in apiService
      onSuccess: () => {
          alert('Rental cancelled successfully!');
          queryClient.invalidateQueries({ queryKey: ['rentalHistory'] }); // Refetch history
          queryClient.invalidateQueries({ queryKey: ['availableCars'] }); // Refetch cars
      },
      onError: (err) => {
          alert(`Failed to cancel rental: ${err.message}`);
      }
  });

   const { mutate: postReview, isPending: isReviewing } = useMutation({
      mutationFn: submitReview, // Assumes submitReview(rentalId, reviewData) exists
      onSuccess: () => {
          alert('Review submitted successfully!');
          // Optionally refetch history or specific review details
          queryClient.invalidateQueries({ queryKey: ['rentalHistory'] });
      },
      onError: (err) => {
          alert(`Failed to submit review: ${err.message}`);
      }
  });

  const handleCancel = (rentalId) => {
      if (window.confirm("Are you sure you want to cancel this rental?")) {
          cancelBooking(rentalId);
      }
  };

  const handleReview = (rentalId) => {
      // TODO: Show a modal or form to get review data (rating, comment)
      const reviewData = { rating: 5, comment: "Great car!" }; // Example data
      postReview({ rentalId, reviewData }); // Pass as object if needed by mutationFn
  };
  */

  return (
    <div>
      <h2>My Rental History</h2>
      {isLoading && <LoadingSpinner />}
      <ErrorMessage message={isError ? error?.message : null} />

      {!isLoading && rentals.length === 0 && !isError && (
        <p>No rental history available.</p>
      )}

      {!isLoading && rentals.length > 0 && (
        <ul style={{ listStyle: 'none', padding: 0 }}>
          {rentals.map((rental) => (
            <li key={rental.id} style={{ border: '1px solid #ddd', marginBottom: '10px', padding: '15px', borderRadius: '5px' }}>
              <p><strong>Rental ID:</strong> {rental.id}</p>
              <p><strong>Car ID:</strong> {rental.car_id} {/* TODO: Display car details? */}</p>
              <p><strong>Pickup:</strong> {new Date(rental.pickup_datetime).toLocaleString()}</p>
              <p><strong>Dropoff:</strong> {new Date(rental.dropoff_datetime).toLocaleString()}</p>
              <p><strong>Status:</strong> <span style={{ fontWeight: 'bold', color: rental.status === 'Returned' ? 'green' : rental.status === 'Cancelled' ? 'red' : 'inherit' }}>{rental.status}</span></p>
              <p><strong>Booked on:</strong> {rental.booking_date ? new Date(rental.booking_date).toLocaleDateString() : new Date(rental.created_at).toLocaleDateString()}</p>
              <div style={{marginTop: '10px'}}>
                  {(rental.status === 'Booked' || rental.status === 'Confirmed') && (
                    // <button onClick={() => handleCancel(rental.id)} disabled={isCancelling} style={{ marginRight: '10px' }}>{isCancelling ? 'Cancelling...' : 'Cancel Rental'}</button>
                     <button disabled style={{ marginRight: '10px' }}>Cancel Rental (WIP)</button> // Placeholder
                  )}
                  {rental.status === 'Returned' && (
                    // <button onClick={() => handleReview(rental.id)} disabled={isReviewing}>{isReviewing ? 'Submitting...' : 'Submit Review'}</button>
                    <button disabled>Submit Review (WIP)</button> // Placeholder
                  )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default RentalHistory;