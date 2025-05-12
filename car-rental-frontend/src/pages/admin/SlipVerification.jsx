import React, { useState, useMemo, useCallback, useContext } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext';
import { fetchRentalsPendingVerification, verifyPaymentSlip } from '../../services/apiService';
import LoadingSpinner from '../../components/LoadingSpinner';
import ErrorMessage from '../../components/ErrorMessage';
import SlipPreviewModal from '../../components/admin/SlipPreviewModal'; // <<< IMPORT MODAL
import '../../components/admin/AdminCommon.css';

const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const SlipVerification = () => {
    const queryClient = useQueryClient();
    const { logout } = useContext(AuthContext);
    const navigate = useNavigate();
    const [sortConfig, setSortConfig] = useState({ key: 'payment_date', direction: 'ascending' });
    const [isModalOpen, setIsModalOpen] = useState(false); // <<< STATE FOR MODAL
    const [selectedRentalForModal, setSelectedRentalForModal] = useState(null); // <<< STATE FOR MODAL DATA

    const { data: pendingRentals = [], isLoading, isError, error: queryError } = useQuery({
        queryKey: ['rentalsPendingVerification'],
        queryFn: fetchRentalsPendingVerification,
        staleTime: 1000 * 60,
        refetchInterval: 1000 * 60 * 2,
    });

    const sortedPendingRentals = useMemo(() => {
        let sortableItems = [...pendingRentals];
        if (sortConfig.key !== null) {
            sortableItems.sort((a, b) => {
                const aValue = a[sortConfig.key] ?? '';
                const bValue = b[sortConfig.key] ?? '';
                if (sortConfig.key === 'rental_id' || sortConfig.key === 'payment_amount') {
                    const numA = parseFloat(aValue) || 0;
                    const numB = parseFloat(bValue) || 0;
                    if (numA < numB) return sortConfig.direction === 'ascending' ? -1 : 1;
                    if (numA > numB) return sortConfig.direction === 'ascending' ? 1 : -1;
                    return 0;
                } else if (sortConfig.key === 'payment_date') {
                    const dateA = new Date(aValue);
                    const dateB = new Date(bValue);
                    if (dateA < dateB) return sortConfig.direction === 'ascending' ? -1 : 1;
                    if (dateA > dateB) return sortConfig.direction === 'ascending' ? 1 : -1;
                    return 0;
                } else {
                    let strA = '';
                    let strB = '';
                    if (sortConfig.key === 'customer_name') {
                        strA = String(a.customer_name).toLowerCase();
                        strB = String(b.customer_name).toLowerCase();
                    } else if (sortConfig.key === 'car') {
                         strA = `${String(a.car_brand)} ${String(a.car_model)}`.toLowerCase();
                         strB = `${String(b.car_brand)} ${String(b.car_model)}`.toLowerCase();
                    } else {
                         strA = String(aValue).toLowerCase();
                         strB = String(bValue).toLowerCase();
                    }
                    if (strA < strB) return sortConfig.direction === 'ascending' ? -1 : 1;
                    if (strA > strB) return sortConfig.direction === 'ascending' ? 1 : -1;
                    return 0;
                }
            });
        }
        return sortableItems;
    }, [pendingRentals, sortConfig]);

    const requestSort = useCallback((key) => {
        let direction = 'ascending';
        if (sortConfig.key === key && sortConfig.direction === 'ascending') {
            direction = 'descending';
        }
        setSortConfig({ key, direction });
    }, [sortConfig]);

    const { mutate: handleVerification, isPending: isVerifying, variables: verifyingVariables, error: mutationError } = useMutation({
        mutationFn: ({ rentalId, isApproved }) => verifyPaymentSlip(rentalId, isApproved),
        onSuccess: (data, { rentalId, isApproved }) => {
            alert(`Rental ${rentalId} payment has been ${isApproved ? 'approved' : 'rejected'}.`);
            setIsModalOpen(false); // <<< CLOSE MODAL ON SUCCESS
            setSelectedRentalForModal(null);
            queryClient.invalidateQueries({ queryKey: ['rentalsPendingVerification'] });
            queryClient.invalidateQueries({ queryKey: ['rentals'] });
            queryClient.invalidateQueries({ queryKey: ['adminDashboardData'] });
        },
        onError: (err, { rentalId }) => {
            console.error(`Mutation error verifying rental ${rentalId}:`, err);
            // Error handling in modal or here
            if (err.response && err.response.status === 401) {
                alert("Your session has expired. Please log in again.");
                logout();
                navigate('/admin/login', { replace: true });
                setIsModalOpen(false);
                setSelectedRentalForModal(null);
            }
             // setError for modal or keep it global
        }
    });

    const openSlipModal = (rental) => {
        if (!rental.slip_url) {
            alert("No slip available for this rental.");
            return;
        }
        setSelectedRentalForModal(rental);
        setIsModalOpen(true);
    };

    const closeSlipModal = () => {
        setIsModalOpen(false);
        setSelectedRentalForModal(null);
    };

    const onApproveInModal = (rentalId) => {
        handleVerification({ rentalId, isApproved: true });
    };

    const onRejectInModal = (rentalId) => {
        handleVerification({ rentalId, isApproved: false });
    };

    const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };

    return (
        <div className="admin-container">
            <div className="admin-header">
                <h2>Payment Slip Verification</h2>
            </div>
            <ErrorMessage message={queryError?.message || (mutationError && mutationError.response?.status !== 401 ? mutationError.message : null)} />
            {isLoading && <LoadingSpinner />}
            {!isLoading && !isError && (
                <div className="admin-table-wrapper">
                    <table className="admin-table">
                        <thead>
                             <tr>
                                <th style={thSortableStyle} onClick={() => requestSort('rental_id')}>Rental ID <SortIcon direction={sortConfig.key === 'rental_id' ? sortConfig.direction : null} /></th>
                                <th style={thSortableStyle} onClick={() => requestSort('customer_name')}>Customer <SortIcon direction={sortConfig.key === 'customer_name' ? sortConfig.direction : null} /></th>
                                <th style={thSortableStyle} onClick={() => requestSort('car')}>Car <SortIcon direction={sortConfig.key === 'car' ? sortConfig.direction : null} /></th>
                                <th style={thSortableStyle} onClick={() => requestSort('payment_amount')}>Amount <SortIcon direction={sortConfig.key === 'payment_amount' ? sortConfig.direction : null} /></th>
                                <th>Slip</th>
                                <th style={thSortableStyle} onClick={() => requestSort('payment_date')}>Submitted <SortIcon direction={sortConfig.key === 'payment_date' ? sortConfig.direction : null} /></th>
                                {/* Removed direct action buttons from table row, will be in modal */}
                            </tr>
                        </thead>
                         <tbody>
                            {sortedPendingRentals.length === 0 ? (
                                <tr className="admin-table-placeholder">
                                    <td colSpan="6">No payments pending verification.</td>
                                </tr>
                            ) : (
                                sortedPendingRentals.map((rental) => {
                                    return (
                                        <tr key={rental.rental_id}>
                                            <td>{rental.rental_id}</td>
                                            <td>{rental.customer_name} (ID: {rental.customer_id})</td>
                                            <td>{rental.car_brand} {rental.car_model} (ID: {rental.car_id})</td>
                                            <td>฿{rental.payment_amount.toFixed(2)}</td>
                                            <td>
                                                {rental.slip_url ? (
                                                    <button
                                                        onClick={() => openSlipModal(rental)}
                                                        className="admin-button admin-button-info admin-button-sm"
                                                    >
                                                        View & Verify Slip
                                                    </button>
                                                ) : (
                                                    <span style={{color: 'grey', fontStyle: 'italic'}}>No Slip</span>
                                                )}
                                            </td>
                                            <td>{new Date(rental.payment_date).toLocaleString()}</td>
                                        </tr>
                                    );
                                })
                            )}
                        </tbody>
                    </table>
                </div>
            )}
            {isModalOpen && selectedRentalForModal && (
                <SlipPreviewModal
                    rental={selectedRentalForModal}
                    onClose={closeSlipModal}
                    onApprove={onApproveInModal}
                    onReject={onRejectInModal}
                    isVerifying={isVerifying && verifyingVariables?.rentalId === selectedRentalForModal.rental_id}
                />
            )}
        </div>
    );
};

export default SlipVerification;