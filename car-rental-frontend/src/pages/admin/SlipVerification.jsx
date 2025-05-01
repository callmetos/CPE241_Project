import React, { useState, useMemo, useCallback, useContext } from 'react'; // Import hooks
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext';
import { fetchRentalsPendingVerification, verifyPaymentSlip } from '../../services/apiService';
import LoadingSpinner from '../../components/LoadingSpinner';
import ErrorMessage from '../../components/ErrorMessage';
import '../../components/admin/AdminCommon.css';

// --- Sort Icon Component (Optional) ---
const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const SlipVerification = () => {
    const queryClient = useQueryClient();
    const { logout } = useContext(AuthContext);
    const navigate = useNavigate();
    // --- State for sorting ---
    const [sortConfig, setSortConfig] = useState({ key: 'payment_date', direction: 'ascending' }); // Default sort by submission date asc

    // Fetch rentals pending verification
    const { data: pendingRentals = [], isLoading, isError, error: queryError } = useQuery({
        queryKey: ['rentalsPendingVerification'],
        queryFn: fetchRentalsPendingVerification,
        staleTime: 1000 * 60,
        refetchInterval: 1000 * 60 * 2,
    });

     // --- Sorting Logic using useMemo ---
     const sortedPendingRentals = useMemo(() => {
        let sortableItems = [...pendingRentals];
        if (sortConfig.key !== null) {
            sortableItems.sort((a, b) => {
                // Use the correct field names from RentalPendingVerification struct
                const aValue = a[sortConfig.key] ?? '';
                const bValue = b[sortConfig.key] ?? '';

                if (sortConfig.key === 'rental_id' || sortConfig.key === 'payment_amount') {
                    // Numeric comparison
                    const numA = parseFloat(aValue) || 0;
                    const numB = parseFloat(bValue) || 0;
                    if (numA < numB) return sortConfig.direction === 'ascending' ? -1 : 1;
                    if (numA > numB) return sortConfig.direction === 'ascending' ? 1 : -1;
                    return 0;
                } else if (sortConfig.key === 'payment_date') {
                    // Date comparison
                    const dateA = new Date(aValue);
                    const dateB = new Date(bValue);
                    if (dateA < dateB) return sortConfig.direction === 'ascending' ? -1 : 1;
                    if (dateA > dateB) return sortConfig.direction === 'ascending' ? 1 : -1;
                    return 0;
                } else {
                    // String comparison (case-insensitive) for customer_name, car_brand, car_model
                    let strA = '';
                    let strB = '';
                    if (sortConfig.key === 'customer_name') {
                        strA = String(a.customer_name).toLowerCase();
                        strB = String(b.customer_name).toLowerCase();
                    } else if (sortConfig.key === 'car') { // Combine brand and model for sorting
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

    // --- Request Sort Function ---
    const requestSort = useCallback((key) => {
        let direction = 'ascending';
        if (sortConfig.key === key && sortConfig.direction === 'ascending') {
            direction = 'descending';
        }
        setSortConfig({ key, direction });
    }, [sortConfig]);

    // Mutation for Approve/Reject
    const { mutate: handleVerification, isPending: isVerifying, variables: verifyingVariables, error: mutationError } = useMutation({
        mutationFn: ({ rentalId, isApproved }) => verifyPaymentSlip(rentalId, isApproved),
        onSuccess: (data, { rentalId, isApproved }) => {
            alert(`Rental ${rentalId} payment has been ${isApproved ? 'approved' : 'rejected'}.`);
            queryClient.invalidateQueries({ queryKey: ['rentalsPendingVerification'] });
            queryClient.invalidateQueries({ queryKey: ['rentals'] });
            queryClient.invalidateQueries({ queryKey: ['adminDashboardData'] });
        },
        onError: (err, { rentalId }) => {
            console.error(`Mutation error verifying rental ${rentalId}:`, err);
            if (err.response && err.response.status === 401) {
                alert("Your session has expired. Please log in again.");
                logout();
                navigate('/admin/login', { replace: true });
            }
            // Other errors will be displayed by ErrorMessage component below
        }
    });

    // Handler for button clicks
    const onVerify = (rentalId, approve) => {
        const action = approve ? "approve" : "reject";
        if (window.confirm(`Are you sure you want to ${action} the payment for rental ID ${rentalId}?`)) {
            handleVerification({ rentalId, isApproved: approve });
        }
    };

    // Render slip link/button
    const renderSlip = (slipUrl) => {
       if (!slipUrl) {
           return <span style={{color: 'grey', fontStyle: 'italic'}}>No Slip</span>;
       }
       // Assuming slipUrl from backend is relative like "/uploads/slips/..."
       // If it's absolute, no need to prepend API_URL or base path
       const fullSlipUrl = slipUrl.startsWith('http') ? slipUrl : `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}${slipUrl}`;
       return (
           <a href={fullSlipUrl} target="_blank" rel="noopener noreferrer" className="admin-button admin-button-info admin-button-sm">
               View Slip
           </a>
       );
    };

    // Style for sortable header
    const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };

    return (
        <div className="admin-container">
            <div className="admin-header">
                <h2>Payment Slip Verification</h2>
            </div>

            {/* Display errors from query or non-401 mutation errors */}
            <ErrorMessage message={queryError?.message || (mutationError && mutationError.response?.status !== 401 ? mutationError.message : null)} />

            {isLoading && <LoadingSpinner />}

            {!isLoading && !isError && (
                <div className="admin-table-wrapper">
                    <table className="admin-table">
                        <thead>
                             <tr>
                                {/* Sortable Headers */}
                                <th style={thSortableStyle} onClick={() => requestSort('rental_id')}>
                                    Rental ID <SortIcon direction={sortConfig.key === 'rental_id' ? sortConfig.direction : null} />
                                </th>
                                <th style={thSortableStyle} onClick={() => requestSort('customer_name')}>
                                    Customer <SortIcon direction={sortConfig.key === 'customer_name' ? sortConfig.direction : null} />
                                </th>
                                <th style={thSortableStyle} onClick={() => requestSort('car')}> {/* Sort by combined car string */}
                                    Car <SortIcon direction={sortConfig.key === 'car' ? sortConfig.direction : null} />
                                </th>
                                <th style={thSortableStyle} onClick={() => requestSort('payment_amount')}>
                                    Amount <SortIcon direction={sortConfig.key === 'payment_amount' ? sortConfig.direction : null} />
                                </th>
                                <th>Slip</th>
                                <th style={thSortableStyle} onClick={() => requestSort('payment_date')}>
                                    Submitted <SortIcon direction={sortConfig.key === 'payment_date' ? sortConfig.direction : null} />
                                </th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                         <tbody>
                            {/* Use sortedPendingRentals */}
                            {sortedPendingRentals.length === 0 ? (
                                <tr className="admin-table-placeholder">
                                    <td colSpan="7">No payments pending verification.</td>
                                </tr>
                            ) : (
                                sortedPendingRentals.map((rental) => {
                                    const isCurrentVerifying = isVerifying && verifyingVariables?.rentalId === rental.rental_id;
                                    return (
                                        <tr key={rental.rental_id}>
                                            <td>{rental.rental_id}</td>
                                            <td>{rental.customer_name} (ID: {rental.customer_id})</td>
                                            <td>{rental.car_brand} {rental.car_model} (ID: {rental.car_id})</td>
                                            <td>฿{rental.payment_amount.toFixed(2)}</td>
                                            <td>{renderSlip(rental.slip_url)}</td>
                                            <td>{new Date(rental.payment_date).toLocaleString()}</td>
                                            <td className="actions admin-action-buttons">
                                                <button
                                                    onClick={() => onVerify(rental.rental_id, true)}
                                                    className="admin-button admin-button-success admin-button-sm"
                                                    disabled={isCurrentVerifying}
                                                >
                                                    {isCurrentVerifying && verifyingVariables?.isApproved === true ? '...' : 'Approve'}
                                                </button>
                                                <button
                                                    onClick={() => onVerify(rental.rental_id, false)}
                                                    className="admin-button admin-button-danger admin-button-sm"
                                                    disabled={isCurrentVerifying}
                                                >
                                                    {isCurrentVerifying && verifyingVariables?.isApproved === false ? '...' : 'Reject'}
                                                </button>
                                            </td>
                                        </tr>
                                    );
                                })
                            )}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
};

export default SlipVerification;
