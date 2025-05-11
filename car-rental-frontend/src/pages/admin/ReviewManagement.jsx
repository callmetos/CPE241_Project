import React, { useState, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAllReviewsAdmin, deleteReview } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import '../../components/admin/AdminCommon.css'; // Import common admin styles

const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const StarRatingDisplay = ({ rating }) => {
    const stars = [];
    for (let i = 1; i <= 5; i++) {
        stars.push(
            <span key={i} style={{ color: i <= rating ? '#ffc107' : '#e0e0e0', fontSize: '1em' }}>
                ★
            </span>
        );
    }
    return <div style={{ whiteSpace: 'nowrap' }}>{stars}</div>;
};


const ReviewManagement = () => {
    const queryClient = useQueryClient();
    const [error, setError] = useState('');
    const [uiFilters, setUiFilters] = useState({
        rating: '',
        customer_id: '',
        car_id: '',
        keyword: '',
    });
    const [sortConfig, setSortConfig] = useState({ key: 'review_created_at', direction: 'DESC' });
    const [currentPage, setCurrentPage] = useState(1);
    const [itemsPerPage, setItemsPerPage] = useState(10);

    const activeQueryParams = useMemo(() => {
        const params = {
            page: currentPage,
            limit: itemsPerPage,
            sort_by: sortConfig.key,
            sort_dir: sortConfig.direction.toUpperCase(),
        };
        if (uiFilters.rating) params.rating = parseInt(uiFilters.rating);
        if (uiFilters.customer_id) params.customer_id = parseInt(uiFilters.customer_id);
        if (uiFilters.car_id) params.car_id = parseInt(uiFilters.car_id);
        if (uiFilters.keyword) params.keyword = uiFilters.keyword;
        return params;
    }, [uiFilters, currentPage, itemsPerPage, sortConfig]);

    const {
        data: reviewsData,
        isLoading: isLoadingReviews,
        isError: isErrorReviews,
        error: reviewsError,
    } = useQuery({
        queryKey: ['adminReviews', activeQueryParams],
        queryFn: () => fetchAllReviewsAdmin(activeQueryParams),
        staleTime: 1000 * 60 * 1, // Cache for 1 minute
        keepPreviousData: true,
    });

    const currentReviewsToDisplay = reviewsData?.reviews || [];
    const totalItems = reviewsData?.total_count || 0;
    const totalPages = Math.ceil(totalItems / itemsPerPage);

    const requestSort = useCallback((key) => {
        let direction = 'ASC';
        if (sortConfig.key === key && sortConfig.direction === 'ASC') {
            direction = 'DESC';
        }
        setSortConfig({ key, direction });
        setCurrentPage(1);
    }, [sortConfig]);

    const { mutate: removeReview, isPending: isDeleting } = useMutation({
        mutationFn: deleteReview,
        onSuccess: (data, reviewId) => {
            setError('');
            queryClient.invalidateQueries({ queryKey: ['adminReviews'] });
            alert(`Review ID ${reviewId} deleted successfully!`);
        },
        onError: (err, reviewId) => {
            setError(`Failed to delete review ${reviewId}: ${err.message}`);
        },
    });

    const handleDeleteReview = (reviewId) => {
        if (window.confirm(`Are you sure you want to delete review ID ${reviewId}? This action cannot be undone.`)) {
            setError('');
            removeReview(reviewId);
        }
    };

    const handleFilterChange = (e) => {
        const { name, value } = e.target;
        setUiFilters(prev => ({ ...prev, [name]: value }));
        setCurrentPage(1);
    };

    const handleClearFilters = () => {
        setUiFilters({ rating: '', customer_id: '', car_id: '', keyword: '' });
        setCurrentPage(1);
    };

    const paginate = (pageNumber) => {
        if (pageNumber > 0 && pageNumber <= totalPages) {
            setCurrentPage(pageNumber);
        }
    };

    const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };
    const filterSectionStyle = { display: 'flex', flexWrap: 'wrap', gap: '15px', padding: '15px', marginBottom: '20px', backgroundColor: '#f8f9fa', borderRadius: '4px', border: '1px solid var(--admin-border-color)' };
    const filterGroupStyle = { display: 'flex', flexDirection: 'column', flex: '1 1 150px' };
    const filterLabelStyle = { marginBottom: '5px', fontSize: '0.85em', fontWeight: '500', color: 'var(--admin-text-medium)' };
    const filterInputStyle = { padding: '8px', fontSize: '0.9em', border: '1px solid #ccc', borderRadius: '4px' };
    const filterButtonStyle = { alignSelf: 'flex-end', padding: '8px 15px', minHeight: '36px' };
    const paginationContainerStyle = { display: 'flex', justifyContent: 'center', alignItems: 'center', marginTop: '20px', paddingTop: '15px', borderTop: '1px solid #eee' };
    const paginationButtonStyle = (isActive) => ({ margin: '0 5px', padding: '8px 12px', cursor: 'pointer', backgroundColor: isActive ? 'var(--admin-primary)' : '#f0f0f0', color: isActive ? 'white' : '#333', border: `1px solid ${isActive ? 'var(--admin-primary)' : '#ccc'}`, borderRadius: '4px', fontWeight: isActive ? 'bold' : 'normal', });
    const paginationNavButtonStyle = { ...paginationButtonStyle(false), backgroundColor: '#e9ecef' };


    return (
        <div className="admin-container">
            <div className="admin-header">
                <h2>Review Management</h2>
            </div>

            <div style={filterSectionStyle}>
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-rating" style={filterLabelStyle}>Rating (1-5)</label>
                    <input type="number" id="filter-rating" name="rating" style={filterInputStyle} value={uiFilters.rating} onChange={handleFilterChange} placeholder="e.g., 5" min="1" max="5"/>
                </div>
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-customer-id" style={filterLabelStyle}>Customer ID</label>
                    <input type="number" id="filter-customer-id" name="customer_id" style={filterInputStyle} value={uiFilters.customer_id} onChange={handleFilterChange} placeholder="e.g., 12"/>
                </div>
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-car-id" style={filterLabelStyle}>Car ID</label>
                    <input type="number" id="filter-car-id" name="car_id" style={filterInputStyle} value={uiFilters.car_id} onChange={handleFilterChange} placeholder="e.g., 3"/>
                </div>
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-keyword" style={filterLabelStyle}>Comment Keyword</label>
                    <input type="text" id="filter-keyword" name="keyword" style={filterInputStyle} value={uiFilters.keyword} onChange={handleFilterChange} placeholder="e.g., excellent"/>
                </div>
                 <div style={{...filterGroupStyle, flexDirection: 'row', alignItems: 'flex-end', gap: '10px', flexBasis: 'auto' }}>
                    <button onClick={handleClearFilters} className="admin-button admin-button-secondary" style={filterButtonStyle} disabled={isLoadingReviews}>Clear Filters</button>
                </div>
            </div>

            <ErrorMessage message={error || (isErrorReviews ? `Error: ${reviewsError?.message}` : null)} />
            {isLoadingReviews && <LoadingSpinner />}

            {!isLoadingReviews && (
                <>
                    <div className="admin-table-wrapper">
                        <table className="admin-table">
                            <thead>
                                <tr>
                                    <th style={thSortableStyle} onClick={() => requestSort('id')}>ID <SortIcon direction={sortConfig.key === 'id' ? sortConfig.direction.toLowerCase() : null} /></th>
                                    <th style={thSortableStyle} onClick={() => requestSort('review_created_at')}>Date <SortIcon direction={sortConfig.key === 'review_created_at' ? sortConfig.direction.toLowerCase() : null} /></th>
                                    <th style={thSortableStyle} onClick={() => requestSort('rating')}>Rating <SortIcon direction={sortConfig.key === 'rating' ? sortConfig.direction.toLowerCase() : null} /></th>
                                    <th>Car (ID)</th>
                                    <th>Customer (ID)</th>
                                    <th className="wrap-text" style={{minWidth: '250px'}}>Comment</th>
                                    <th>Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {currentReviewsToDisplay.length === 0 ? (
                                    <tr className="admin-table-placeholder"><td colSpan="7">No reviews found matching criteria.</td></tr>
                                ) : (
                                    currentReviewsToDisplay.map((review) => (
                                        <tr key={review.id}>
                                            <td>{review.id}</td>
                                            <td>{new Date(review.review_created_at).toLocaleDateString()}</td>
                                            <td><StarRatingDisplay rating={review.rating} /></td>
                                            <td>{review.car_brand} {review.car_model} ({review.car_id})</td>
                                            <td>{review.customer_name || 'N/A'} ({review.customer_id})</td>
                                            <td className="wrap-text" style={{maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'normal'}}>
                                                {review.comment || <span style={{color: 'grey'}}>No comment</span>}
                                            </td>
                                            <td className="actions admin-action-buttons">
                                                <button
                                                    onClick={() => handleDeleteReview(review.id)}
                                                    className="admin-button admin-button-danger admin-button-sm"
                                                    disabled={isDeleting && updatingVariables === review.id}
                                                >
                                                    {isDeleting && updatingVariables === review.id ? '...' : 'Delete'}
                                                </button>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
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
        </div>
    );
};

export default ReviewManagement;
