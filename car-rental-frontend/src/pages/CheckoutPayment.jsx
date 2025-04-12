import React, { useState, useEffect } from 'react';
import { useParams, useLocation, useNavigate, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
//import { QRCode } from "qrcode.react"; // Use correct import
import { checkPaymentStatus } from '../services/apiService';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';

//Assume these components are imported from a shared location
// import { PriceDetailsDisplay, RentalInfoDisplay, CarSummaryDisplay } from '../components/checkout/SummaryDisplays';
// Using placeholders for now:
const PriceDetailsDisplay = ({ priceData }) => priceData ? <div style={{border:'1px dashed gray', padding: '5px', margin:'5px'}}>Price: {priceData.total?.toFixed(2)} {priceData.currency} (Placeholder)</div> : <div style={{border:'1px dashed gray', padding: '5px', margin:'5px'}}>Loading Price...</div>;
const RentalInfoDisplay = ({ rentalData }) => rentalData ? <div style={{border:'1px dashed gray', padding: '5px', margin:'5px'}}>Rental: {new Date(rentalData.pickup_datetime).toLocaleDateString()} (Placeholder)</div> : null;
const CarSummaryDisplay = ({ carData }) => carData ? <div style={{border:'1px dashed gray', padding: '5px', margin:'5px'}}>Car: {carData.brand} {carData.model} (Placeholder)</div> : null;

const POLLING_INTERVAL = 5000;

const CheckoutPayment = () => {
    const { rentalId } = useParams(); const location = useLocation(); const navigate = useNavigate();
    const { paymentDetails, rental, price } = location.state || {};
    const [isPolling, setIsPolling] = useState(false); const [paymentError, setPaymentError] = useState(''); const [finalStatus, setFinalStatus] = useState(paymentDetails?.status);

    useQuery({
        queryKey: ['paymentStatus', paymentDetails?.paymentId], queryFn: () => checkPaymentStatus(paymentDetails.paymentId),
        enabled: isPolling && !!paymentDetails?.paymentId && finalStatus === 'Pending', refetchInterval: POLLING_INTERVAL,
        refetchIntervalInBackground: true, refetchOnWindowFocus: true, retry: false,
        onSuccess: (data) => { if (data?.status && data.status !== 'Pending') { setIsPolling(false); setFinalStatus(data.status); if (data.status === 'Paid') { alert("Payment successful!"); navigate(`/rental-history`, { replace: true }); } else { setPaymentError(`Payment ${data.status}.`); }}},
        onError: (error) => { setIsPolling(false); setPaymentError(`Error checking status: ${error.message}.`); }
    });

    useEffect(() => { if (paymentDetails?.paymentId && paymentDetails.status === 'Pending') { setFinalStatus('Pending'); setIsPolling(true); } else if (paymentDetails?.status) { setFinalStatus(paymentDetails.status); setIsPolling(false); } return () => setIsPolling(false); }, [paymentDetails]);

    const handleBack = () => navigate(`/checkout/${rentalId}/user-info`);

    if (!paymentDetails || !rental || !price) { return ( <div style={{ padding: '20px' }}> <h2>Error</h2> <ErrorMessage message="Checkout info missing." /> <Link to="/">Home</Link> </div> ); }

    const isQRCodeMethod = paymentDetails.method === 'QR Code' || paymentDetails.method === 'PromptPay';
    const paymentAmount = paymentDetails.amount || price?.total; const paymentCurrency = paymentDetails.currency || price?.currency;
    const qrCodeContainerStyle = { textAlign: 'center', padding: '20px', border: '1px solid #ddd', borderRadius: '10px', margin: '20px 0', backgroundColor: '#fff' }; const instructionStyle = { marginTop: '10px', color: '#555', fontSize: '0.9em' }; const statusStyle = (status) => ({ marginTop: '15px', fontWeight: 'bold', color: status === 'Paid' ? 'green' : status === 'Failed' || status === 'Cancelled' ? 'red' : 'orange' });

    return ( <div style={{ maxWidth: '900px', margin: '20px auto', padding: '20px', border: '1px solid #eee', borderRadius: '10px', backgroundColor:'#fff' }}> <h2 style={{ textAlign: 'center', marginBottom: '20px' }}>Complete Your Payment</h2> <div style={{textAlign: 'center', marginBottom: '20px', color: 'gray'}}>Step 1 &gt; Step 2 &gt; Step 3 &gt; <strong style={{color: 'black'}}>Step 4 (Payment)</strong></div> <ErrorMessage message={paymentError} /> <div style={{ display: 'flex', gap: '30px', flexWrap: 'wrap' }}> <div style={{ flex: '2 1 400px' }}> <h4>Payment Method: {paymentDetails.method}</h4> <p>Amount Due: <strong>{paymentAmount?.toFixed(2)} {paymentCurrency}</strong></p> {isQRCodeMethod && paymentDetails.qrCodeData && (finalStatus === 'Pending' || isPolling) && ( <div style={qrCodeContainerStyle}> <p>Scan QR code below with your banking app.</p> <div style={{ margin: '15px 0' }}> {paymentDetails.qrCodeData && <QRCode value={paymentDetails.qrCodeData} size={230} level="M" />} </div> <p style={instructionStyle}>Payment Reference (Payment ID): {paymentDetails.paymentId}</p> {isPolling && <div><LoadingSpinner /><span style={{verticalAlign: 'middle', marginLeft:'5px'}}>Checking payment status...</span></div>} </div> )} {!isPolling && finalStatus && finalStatus !== 'Pending' && <p style={statusStyle(finalStatus)}>Payment Status: {finalStatus}</p>} {!isPolling && finalStatus === 'Pending' && <p style={statusStyle(finalStatus)}>Payment Status: {finalStatus}</p>} {paymentDetails.method === 'Bank transfers' && ( <div style={{marginTop: '20px', padding: '15px', border: '1px dashed #ccc'}}> <p><strong>Instructions:</strong></p> <p>{paymentDetails.instructions || `Please transfer ${paymentAmount?.toFixed(2)} ${paymentCurrency} and use Payment ID ${paymentDetails.paymentId} as reference.`}</p> {finalStatus === 'Pending' && <p style={instructionStyle}>After completing the transfer, please wait for confirmation.</p>} </div> )} </div> <div style={{ flex: '1 1 250px', borderLeft: '1px solid #eee', paddingLeft: '30px' }}> <CarSummaryDisplay carData={rental?.car} /> <RentalInfoDisplay rentalData={rental} /> <PriceDetailsDisplay priceData={price} /> </div> </div> <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '30px', borderTop: '1px solid #eee', paddingTop: '20px' }}> <button onClick={handleBack} disabled={isPolling} style={{ padding: '10px 20px' }}>Back</button> {finalStatus === 'Paid' && <Link to="/rental-history"><button style={{ padding: '10px 20px', backgroundColor: '#28a745', color: 'white' }}>View Rental History</button></Link>} {(finalStatus === 'Failed' || finalStatus === 'Cancelled') && <button onClick={handleBack} style={{ padding: '10px 20px', backgroundColor: '#ffc107', color: 'black' }}>Payment Failed - Go Back</button>} </div> </div> );
};
export default CheckoutPayment;