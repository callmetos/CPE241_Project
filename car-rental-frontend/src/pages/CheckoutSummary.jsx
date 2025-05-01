import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchRentalDetails, fetchRentalPriceDetails } from '../services/apiService.js';
import LoadingSpinner from '../components/LoadingSpinner.jsx';
import ErrorMessage from '../components/ErrorMessage.jsx';
import '../components/admin/AdminCommon.css';


const PriceDetailsDisplay = ({ priceData, isLoading, isError, error }) => {
    if (isLoading) return <p>Loading price...</p>;
    if (isError) return <ErrorMessage message={`Error loading price: ${error?.message}`} />;
    if (!priceData) return <p>Price details unavailable.</p>;

    return (
        <div style={{ border: '1px dashed #eee', padding: '15px', margin: '15px 0', borderRadius: '4px', background: '#f8f9fa' }}>
            <h5 style={{ marginBottom: '10px' }}>Price Breakdown</h5>
            <p style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>Base Price:</span>
                <span>฿{priceData.base_price?.toFixed(2) ?? 'N/A'}</span>
            </p>
            <p style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>VAT (7%):</span>
                <span>฿{priceData.vat?.toFixed(2) ?? 'N/A'}</span>
            </p>
            <hr style={{margin: '10px 0', borderColor: '#ddd'}}/>
            <p style={{ display: 'flex', justifyContent: 'space-between', fontWeight: 'bold', fontSize: '1.1em' }}>
                <span>Total Amount:</span>
                <span>฿{priceData.total?.toFixed(2) ?? 'N/A'}</span>
            </p>
        </div>
    );
};
const RentalInfoDisplay = ({ rentalData }) => rentalData ? <div style={{fontSize: '0.9em', color: '#555'}}><p><strong>Rental ID:</strong> {rentalData.id}</p><p><strong>Pickup:</strong> {new Date(rentalData.pickup_datetime).toLocaleString()}</p><p><strong>Return:</strong> {new Date(rentalData.dropoff_datetime).toLocaleString()}</p></div> : null;
const CarSummaryDisplay = ({ carData }) => carData ? <div style={{textAlign: 'center'}}><p style={{fontWeight: 'bold'}}>{carData.brand} {carData.model}</p>{carData.image_url && <img src={carData.image_url} alt={`${carData.brand} ${carData.model}`} style={{maxWidth: '120px', height: 'auto', marginTop: '5px', borderRadius: '4px'}} onError={(e) => { e.target.onerror = null; e.target.src='https://placehold.co/120x80/eee/ccc?text=Car'; }} />}</div> : null;


const CheckoutSummary = () => {
  const { rentalId } = useParams();
  const navigate = useNavigate();

  const { data: rental, isLoading: isLoadingRental, isError: isErrorRental, error: errorRental } = useQuery({
    queryKey: ['rentalDetails', rentalId],
    queryFn: () => fetchRentalDetails(rentalId),
    enabled: !!rentalId,
    staleTime: 1000 * 60 * 5,
  });

  const { data: price, isLoading: isLoadingPrice, isError: isErrorPrice, error: errorPrice } = useQuery({
    queryKey: ['rentalPrice', rentalId],
    queryFn: () => fetchRentalPriceDetails(rentalId),
    enabled: !!rentalId && !!rental,
    staleTime: 1000 * 60 * 1,
    refetchOnWindowFocus: false,
  });

  const handleNext = () => {
      if (isLoadingPrice || isErrorPrice || !price) {
          alert("Please wait for price details to load or resolve errors.");
          return;
      }
      navigate(`/checkout/${rentalId}/user-info`);
  };

  const handleBack = () => {
      navigate(-1);
  };

  const isLoading = isLoadingRental;
  const queryError = errorRental || errorPrice;

  const gridStyle = { display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '30px' };
  const summaryColStyle = { borderRight: '1px solid #eee', paddingRight: '30px', '@media (maxWidth: 768px)': { borderRight: 'none', paddingRight: 0 } };
  const priceColStyle = { paddingLeft: '30px', '@media (maxWidth: 768px)': { paddingLeft: 0, borderTop: '1px solid #eee', paddingTop: '20px' } };
  const buttonContainerStyle = { display: 'flex', justifyContent: 'space-between', marginTop: '30px', borderTop: '1px solid #eee', paddingTop: '20px' };


  return (
    <div className="admin-container" style={{ maxWidth: '950px' }}>
      <h2 style={{ textAlign: 'center', marginBottom: '25px', color: '#333' }}>Booking Summary & Price</h2>

      {isLoading && <LoadingSpinner />}
      <ErrorMessage message={queryError?.message} />

      {!isLoading && !queryError && rental && (
        <div>
            <div style={gridStyle}>
                <div style={summaryColStyle}>
                    <h4 style={{ marginBottom: '10px', fontWeight: '500' }}>Rental Information</h4>
                    <RentalInfoDisplay rentalData={rental} />
                    {rental.car && (
                        <div style={{marginTop: '20px'}}>
                            <h5 style={{ marginBottom: '10px', fontWeight: '500' }}>Car Details</h5>
                             <CarSummaryDisplay carData={rental.car} />
                        </div>
                    )}
                </div>
                <div style={priceColStyle}>
                     <h4 style={{ marginBottom: '10px', fontWeight: '500' }}>Price</h4>
                    <PriceDetailsDisplay
                        priceData={price}
                        isLoading={isLoadingPrice}
                        isError={isErrorPrice}
                        error={errorPrice}
                    />
                </div>
            </div>

            <div style={buttonContainerStyle}>
                <button onClick={handleBack} className="admin-button admin-button-secondary">
                    Back
                </button>
                <button
                    onClick={handleNext}
                    className="admin-button admin-button-primary"
                    disabled={isLoadingRental || isLoadingPrice || isErrorPrice || !price}
                >
                    Next: Enter Info
                </button>
            </div>
        </div>
      )}
       {!isLoading && !queryError && !rental && <ErrorMessage message="Could not load rental details (ID might be invalid)." />}
    </div>
  );
};
export default CheckoutSummary;