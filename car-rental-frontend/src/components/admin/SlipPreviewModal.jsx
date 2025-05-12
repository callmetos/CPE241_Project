import React from 'react';
import './AdminForm.css';
import './AdminCommon.css';

const SlipPreviewModal = ({ rental, onClose, onApprove, onReject, isVerifying }) => {
    if (!rental || !rental.slip_url) return null;

    let fullSlipUrl;
    if (rental.slip_url.startsWith('http')) {
        fullSlipUrl = rental.slip_url;
    } else if (rental.slip_url.startsWith('/')) {
        const apiUrlFromEnv = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
        try {
            const urlObject = new URL(apiUrlFromEnv);
            const serverOrigin = urlObject.origin;
            fullSlipUrl = `${serverOrigin}${rental.slip_url}`;
        } catch (e) {
            console.error("Error parsing VITE_API_URL to get origin:", e);
            fullSlipUrl = `http://localhost:8080${rental.slip_url}`;
        }
    } else {
        fullSlipUrl = rental.slip_url;
    }

    const modalContentStyle = {
        backgroundColor: 'white',
        padding: '20px',
        borderRadius: '8px',
        width: 'auto',
        maxWidth: '90vw',
        maxHeight: '90vh',
        display: 'flex',
        flexDirection: 'column',
        boxShadow: '0 5px 15px rgba(0,0,0,0.3)',
        position: 'relative',
    };

    const imageContainerStyle = {
        maxHeight: 'calc(80vh - 150px)',
        overflowY: 'auto',
        marginBottom: '15px',
        textAlign: 'center',
        border: '1px solid #eee',
        padding: '10px',
        borderRadius: '4px',
    };

    const imageStyle = {
        maxWidth: '100%',
        maxHeight: '100%',
        height: 'auto',
        display: 'block',
        margin: '0 auto',
    };

    const detailsStyle = {
        marginBottom: '15px',
        fontSize: '0.9em',
        color: '#333',
    };

    const detailItemStyle = { marginBottom: '5px'};

    const buttonContainerStyle = {
        display: 'flex',
        justifyContent: 'space-between',
        marginTop: 'auto',
        paddingTop: '15px',
        borderTop: '1px solid #eee',
    };

    return (
        <div className="admin-form-overlay" onClick={onClose}>
            <div style={modalContentStyle} onClick={(e) => e.stopPropagation()}>
                <h3 className="admin-form-header" style={{textAlign: 'center', marginBottom: '15px'}}>
                    Verify Slip for Rental ID: {rental.rental_id}
                </h3>
                <div style={imageContainerStyle}>
                    <img src={fullSlipUrl} alt={`Slip for Rental ${rental.rental_id}`} style={imageStyle} />
                </div>
                <div style={detailsStyle}>
                    <p style={detailItemStyle}><strong>Customer:</strong> {rental.customer_name} (ID: {rental.customer_id})</p>
                    <p style={detailItemStyle}><strong>Car:</strong> {rental.car_brand} {rental.car_model} (ID: {rental.car_id})</p>
                    <p style={detailItemStyle}><strong>Amount:</strong> ฿{rental.payment_amount?.toFixed(2)}</p>
                    <p style={detailItemStyle}><strong>Submitted:</strong> {new Date(rental.payment_date).toLocaleString()}</p>
                </div>
                <div style={buttonContainerStyle}>
                    <button
                        onClick={() => onApprove(rental.rental_id)}
                        className="admin-button admin-button-success"
                        disabled={isVerifying}
                    >
                        {isVerifying ? '...' : 'Approve'}
                    </button>
                    <button
                        onClick={onClose}
                        className="admin-button admin-button-secondary"
                        disabled={isVerifying}
                        style={{margin: '0 10px'}}
                    >
                        Cancel
                    </button>
                    <button
                        onClick={() => onReject(rental.rental_id)}
                        className="admin-button admin-button-danger"
                        disabled={isVerifying}
                    >
                        {isVerifying ? '...' : 'Reject'}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default SlipPreviewModal;