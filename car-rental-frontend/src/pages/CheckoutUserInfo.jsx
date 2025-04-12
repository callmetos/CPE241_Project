import React, { useState, useContext, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { AuthContext } from '../context/AuthContext';
import { fetchRentalDetails, fetchRentalPriceDetails, initiatePayment } from '../services/apiService';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';

// Reusable Display Components (Ideally imported)
const PriceDetailsDisplay = ({ priceData, isLoading, isError, error }) => { if (isLoading || !priceData) return <p>Loading price...</p>; if(isError) return <ErrorMessage message={error?.message}/>; const detailStyle = { display: 'flex', justifyContent: 'space-between', marginBottom: '5px', fontSize: '0.9em' }; const totalStyle = { ...detailStyle, fontWeight: 'bold', borderTop: '1px solid #ccc', paddingTop: '8px', marginTop: '8px', fontSize: '1em' }; return (<div style={{ backgroundColor: '#f8f9fa', padding: '15px', borderRadius: '8px', marginTop: '15px' }}> <h4 style={{ marginBottom: '10px', borderBottom: '1px solid #eee', paddingBottom: '5px' }}>Price Details</h4> <div style={detailStyle}><span>Car Rental Price:</span> <span>{priceData.base_price?.toFixed(2)} {priceData.currency}</span></div> <div style={detailStyle}><span>Drop off fee:</span> <span>{priceData.drop_off_fee?.toFixed(2)} {priceData.currency}</span></div> <div style={detailStyle}><span>VAT:</span> <span>{priceData.vat?.toFixed(2)} {priceData.currency}</span></div> <div style={totalStyle}><span>Total:</span> <span>{priceData.total?.toFixed(2)} {priceData.currency}</span></div> </div> ); };
const RentalInfoDisplay = ({ rentalData }) => { if (!rentalData) return null; return ( <div style={{ marginBottom: '15px' }}> <h4 style={{ marginBottom: '10px' }}>Rental Information</h4> <p><strong>Pick-up:</strong> {rentalData.pickup_location || 'N/A'} <br /> {new Date(rentalData.pickup_datetime).toLocaleString()}</p> <p style={{marginTop: '5px'}}><strong>Return:</strong> {rentalData.return_location || rentalData.pickup_location || 'N/A'} <br /> {new Date(rentalData.dropoff_datetime).toLocaleString()}</p> </div> ); };
const CarSummaryDisplay = ({ carData }) => { if (!carData) return null; return ( <div style={{ textAlign: 'center', marginBottom: '15px' }}> <h4 style={{ marginBottom: '10px' }}>Car Rental Summary</h4> {carData.image_url && <img src={carData.image_url} alt={`${carData.brand} ${carData.model}`} style={{maxWidth: '150px', height: 'auto', marginBottom: '10px'}} />} <p style={{fontWeight: 'bold'}}>{carData.brand} {carData.model}</p> <p style={{fontSize: '0.9em'}}>Size: {carData.size || 'N/A'}</p> </div> ); };

const CheckoutUserInfo = () => {
    const { rentalId } = useParams(); const navigate = useNavigate(); const { user: loggedInUser } = useContext(AuthContext); const queryClient = useQueryClient();
    const [formData, setFormData] = useState({ email: '', firstName: '', lastName: '', phone: '', drivingLicense: '' });
    const [selectedPaymentMethod, setSelectedPaymentMethod] = useState('QR Code');
    const [formError, setFormError] = useState('');

    const { data: rental, isLoading: isLoadingRental, isError: isErrorRental, error: errorRental } = useQuery({ queryKey: ['rentalDetails', rentalId], queryFn: () => fetchRentalDetails(rentalId), enabled: !!rentalId, staleTime: Infinity });
    const { data: price, isLoading: isLoadingPrice, isError: isErrorPrice, error: errorPrice } = useQuery({ queryKey: ['rentalPrice', rentalId], queryFn: () => fetchRentalPriceDetails(rentalId), enabled: !!rentalId });

    useEffect(() => { if (loggedInUser) { setFormData(prev => ({ ...prev, email: loggedInUser.email || '', firstName: loggedInUser.name?.split(' ')[0] || '', lastName: loggedInUser.name?.split(' ').slice(1).join(' ') || '', phone: loggedInUser.phone || '' })); } }, [loggedInUser]);

    const { mutate: proceedToPayment, isPending: isProcessingPayment, error: paymentInitiationError } = useMutation({
        mutationFn: (variables) => initiatePayment(variables.rentalId, variables.paymentMethod),
        onSuccess: (data) => { navigate(`/checkout/${rentalId}/payment`, { state: { paymentDetails: data, rental: rental, price: price } }); },
        onError: (error) => { setFormError(error.message || "Could not proceed."); }
    });

    const handleInputChange = (e) => { setFormData(prev => ({ ...prev, [e.target.name]: e.target.value })); };
    const handlePaymentMethodChange = (e) => { setSelectedPaymentMethod(e.target.value); };
    const handleSubmit = (e) => {
        e.preventDefault(); setFormError('');
        if (!formData.firstName || !formData.lastName || !formData.phone || !formData.drivingLicense || !formData.email) { setFormError("Please fill in all required fields."); return; }
        // TODO: Optional - mutation to update user info before proceeding?
        proceedToPayment({ rentalId, paymentMethod: selectedPaymentMethod });
    };
    const handleBack = () => navigate(`/checkout/${rentalId}/summary`);
    const isLoading = isLoadingRental || isLoadingPrice;
    const queryError = errorRental || errorPrice;
    const isSubmitting = isProcessingPayment;

    // Styles
    const formStyle = { width: '100%' }; const inputGroupStyle = { marginBottom: '15px' }; const labelStyle = { display: 'block', marginBottom: '5px', fontWeight: 'bold' }; const inputStyle = { width: '100%', padding: '10px', border: '1px solid #ccc', borderRadius: '4px' }; const radioGroupStyle = { display: 'flex', gap: '15px', marginTop: '10px' }; const radioLabelStyle = { display: 'flex', alignItems: 'center', gap: '5px', cursor: 'pointer' }; const paymentSectionStyle = { marginTop: '20px', borderTop: '1px solid #eee', paddingTop: '15px'};

    return ( <div style={{ maxWidth: '900px', margin: '20px auto', padding: '20px', border: '1px solid #eee', borderRadius: '10px', backgroundColor:'#fff' }}> <h2 style={{ textAlign: 'center', marginBottom: '20px' }}>Enter Information & Select Payment</h2> <div style={{textAlign: 'center', marginBottom: '20px', color: 'gray'}}>Step 1 &gt; Step 2 &gt; <strong style={{color: 'black'}}>Step 3 (Info)</strong> &gt; Step 4</div> <ErrorMessage message={queryError?.message || formError || paymentInitiationError?.message} /> {isLoading && <LoadingSpinner />} {!isLoading && !queryError && rental && price && ( <div style={{ display: 'flex', gap: '30px', flexWrap: 'wrap' }}> <div style={{ flex: '2 1 400px' }}> <form onSubmit={handleSubmit} style={formStyle}> <h4>Your Information</h4> <p style={{fontSize: '0.8em', color: 'gray', marginBottom: '15px'}}>Please fill in information with English Language only</p> <div style={inputGroupStyle}> <label htmlFor="email" style={labelStyle}>Email *</label> <input type="email" id="email" name="email" style={inputStyle} value={formData.email} onChange={handleInputChange} required readOnly={!!loggedInUser?.email}/> </div> <div style={{ display: 'flex', gap: '15px' }}> <div style={{ ...inputGroupStyle, flex: 1 }}> <label htmlFor="firstName" style={labelStyle}>First Name *</label> <input type="text" id="firstName" name="firstName" style={inputStyle} value={formData.firstName} onChange={handleInputChange} required disabled={isSubmitting}/> </div> <div style={{ ...inputGroupStyle, flex: 1 }}> <label htmlFor="lastName" style={labelStyle}>Last Name *</label> <input type="text" id="lastName" name="lastName" style={inputStyle} value={formData.lastName} onChange={handleInputChange} required disabled={isSubmitting}/> </div> </div> <div style={inputGroupStyle}> <label htmlFor="phone" style={labelStyle}>Phone *</label> <input type="tel" id="phone" name="phone" style={inputStyle} value={formData.phone} onChange={handleInputChange} required disabled={isSubmitting}/> </div> <div style={inputGroupStyle}> <label htmlFor="drivingLicense" style={labelStyle}>Driving License *</label> <input type="text" id="drivingLicense" name="drivingLicense" style={inputStyle} value={formData.drivingLicense} onChange={handleInputChange} required disabled={isSubmitting}/> </div> <div style={paymentSectionStyle}> <h4>Payment methods</h4> <div style={radioGroupStyle}> <label style={radioLabelStyle}> <input type="radio" name="paymentMethod" value="QR Code" checked={selectedPaymentMethod === 'QR Code'} onChange={handlePaymentMethodChange} disabled={isSubmitting}/> QR Code / Prompt Pay </label> <label style={radioLabelStyle}> <input type="radio" name="paymentMethod" value="Bank transfers" checked={selectedPaymentMethod === 'Bank transfers'} onChange={handlePaymentMethodChange} disabled={isSubmitting}/> Bank transfers </label> </div> </div> <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '30px', borderTop: '1px solid #eee', paddingTop: '20px' }}> <button type="button" onClick={handleBack} disabled={isSubmitting} style={{ padding: '10px 20px' }}>Back</button> <button type="submit" disabled={isSubmitting} style={{ padding: '10px 20px', backgroundColor: '#28a745', color: 'white', border: 'none', borderRadius: '5px'}}> {isSubmitting ? 'Processing...' : 'Proceed to Payment'} </button> </div> </form> </div> <div style={{ flex: '1 1 250px', borderLeft: '1px solid #eee', paddingLeft: '30px' }}> <CarSummaryDisplay carData={rental?.car} /> <RentalInfoDisplay rentalData={rental} /> <PriceDetailsDisplay priceData={price} isLoading={isLoadingPrice} isError={isErrorPrice} error={errorPrice} /> </div> </div> )} </div> );
};
export default CheckoutUserInfo;