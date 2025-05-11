import React, { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchRentalHistory, submitReview } from '../services/apiService';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
import ReviewForm from '../components/ReviewForm';
import '../components/admin/AdminCommon.css'; // For button styles, adjust if needed

const RentalHistory = () => {
  const queryClient = useQueryClient();
  const [showReviewModal, setShowReviewModal] = useState(false);
  const [currentRentalForReview, setCurrentRentalForReview] = useState(null);
  const [reviewError, setReviewError] = useState('');

  // State for pagination
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(10); // Or your preferred number

  const {
    data: rentalHistoryData, // API returns { rentals: [], total_count: X, ... }
    isLoading,
    isError,
    error: fetchError,
  } = useQuery({
    queryKey: ['rentalHistory', currentPage, itemsPerPage], // Add currentPage and itemsPerPage to queryKey
    queryFn: () => fetchRentalHistory({ page: currentPage, limit: itemsPerPage }),
    staleTime: 1000 * 60 * 2, // Cache for 2 minutes
    keepPreviousData: true, // Good for pagination UX
  });

  const rentalsToDisplay = useMemo(() => rentalHistoryData?.rentals || [], [rentalHistoryData]);
  const totalItems = rentalHistoryData?.total_count || 0;
  const totalPages = Math.ceil(totalItems / itemsPerPage);

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

  const paginate = (pageNumber) => {
    if (pageNumber > 0 && pageNumber <= totalPages) {
        setCurrentPage(pageNumber);
        window.scrollTo(0,0); // Scroll to top on page change
    }
  };

  const pageContainerStyle = { maxWidth: '1000px', margin: '20px auto', padding: '20px' };
  const listStyle = { listStyle: 'none', padding: 0 };
  const listItemStyle = { border: '1px solid #e0e0e0', marginBottom: '15px', padding: '20px', borderRadius: '8px', backgroundColor: '#fff', boxShadow: '0 2px 4px rgba(0,0,0,0.05)' };
  const detailStyle = { margin: '5px 0', color: '#333', fontSize: '0.95em' };
  const statusStyle = (status) => {
        let backgroundColor; let textColor = 'white';
        switch (status) {
            case 'Returned': backgroundColor = '#198754'; break;
            case 'Cancelled': backgroundColor = '#dc3545'; break;
            case 'Confirmed': backgroundColor = '#0dcaf0'; textColor = '#000'; break;
            case 'Active': backgroundColor = '#0d6efd'; break;
            case 'Booked': backgroundColor = '#ffc107'; textColor = '#333'; break;
            case 'Pending': backgroundColor = '#fd7e14'; break;
            case 'Pending Verification': backgroundColor = '#6f42c1'; break;
            default: backgroundColor = '#6c757d';
        }
        return { fontWeight: 'bold', padding: '3px 8px', borderRadius: '12px', fontSize: '0.85em', backgroundColor: backgroundColor, color: textColor, display: 'inline-block' };
   };
   const actionButtonStyle = { marginRight: '10px', marginTop: '10px' };

   const paginationContainerStyle = { display: 'flex', justifyContent: 'center', alignItems: 'center', marginTop: '30px', paddingTop: '15px', borderTop: '1px solid #eee' };
   const paginationButtonStyle = (isActive) => ({ margin: '0 5px', padding: '8px 12px', cursor: 'pointer', backgroundColor: isActive ? '#007bff' : '#f0f0f0', color: isActive ? 'white' : '#333', border: `1px solid ${isActive ? '#007bff' : '#ccc'}`, borderRadius: '4px', fontWeight: isActive ? 'bold' : 'normal', });
   const paginationNavButtonStyle = { ...paginationButtonStyle(false), backgroundColor: '#e9ecef' };


  return (
    <div style={pageContainerStyle}>
      <h2 style={{ textAlign: 'center', marginBottom: '25px', color: '#333' }}>My Rental History</h2>
      {isLoading && <LoadingSpinner />}
      <ErrorMessage message={isError ? `Error fetching history: ${fetchError?.message}` : null} />

      {!isLoading && rentalsToDisplay.length === 0 && !isError && (
        <p style={{ textAlign: 'center' }}>No rental history available.</p>
      )}

      {!isLoading && rentalsToDisplay.length > 0 && (
        <>
          <ul style={listStyle}>
            {rentalsToDisplay.map((rental) => (
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

          {totalPages > 1 && (
            <div style={paginationContainerStyle}>
                <button onClick={() => paginate(currentPage - 1)} disabled={currentPage === 1} style={paginationNavButtonStyle}>&laquo; Previous</button>
                {Array.from({ length: totalPages }, (_, i) => {
                    const pageNumber = i + 1;
                    const pageRange = 2;
                    const showPage = pageNumber === 1 || pageNumber === totalPages || (pageNumber >= currentPage - pageRange && pageNumber <= currentPage + pageRange);
                    if (showPage) { return (<button key={pageNumber} onClick={() => paginate(pageNumber)} style={paginationButtonStyle(currentPage === pageNumber)}>{pageNumber}</button>); }
                    else if (pageNumber === currentPage - pageRange - 1 || pageNumber === currentPage + pageRange + 1) {
                        if ((pageNumber === 2 && currentPage > 4 && totalPages > 5) || (pageNumber === totalPages - 1 && currentPage < totalPages - 3 && totalPages > 5)) {
                            return <span key={`ellipsis-${pageNumber}`} style={{ margin: '0 5px' }}>...</span>;
                        }
                    }
                    return null;
                })}
                <button onClick={() => paginate(currentPage + 1)} disabled={currentPage === totalPages || totalPages === 0} style={paginationNavButtonStyle}>Next &raquo;</button>
            </div>
          )}
        </>
      )}

      {showReviewModal && currentRentalForReview && (
         <ReviewForm
            rentalId={currentRentalForReview.id}
            carInfo={getCarDisplayName(currentRentalForReview)}
            onSubmit={handleReviewSubmit}
            onClose={handleCloseReviewModal}
            isSubmitting={isReviewing}
            initialError={reviewError} // Pass reviewError to ReviewForm if needed
         />
      )}
    </div>
  );
};

export default RentalHistory;
