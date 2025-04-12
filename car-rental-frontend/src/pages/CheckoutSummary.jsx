import React from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchRentalDetails, fetchRentalPriceDetails } from '../services/apiService';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';

// --- Reusable Display Components (Move to separate files later) ---
const PriceDetailsDisplay = ({ priceData, isLoading, isError, error }) => { if (isLoading || !priceData) return <p>Loading price...</p>; if(isError) return <ErrorMessage message={error?.message}/>; const detailStyle = { display: 'flex', justifyContent: 'space-between', marginBottom: '5px', fontSize: '0.9em' }; const totalStyle = { ...detailStyle, fontWeight: 'bold', borderTop: '1px solid #ccc', paddingTop: '8px', marginTop: '8px', fontSize: '1em' }; return (<div style={{ backgroundColor: '#f8f9fa', padding: '15px', borderRadius: '8px', marginTop: '15px' }}> <h4 style={{ marginBottom: '10px', borderBottom: '1px solid #eee', paddingBottom: '5px' }}>Price Details</h4> <div style={detailStyle}><span>Car Rental Price:</span> <span>{priceData.base_price?.toFixed(2)} {priceData.currency}</span></div> <div style={detailStyle}><span>Drop off fee:</span> <span>{priceData.drop_off_fee?.toFixed(2)} {priceData.currency}</span></div> <div style={detailStyle}><span>VAT:</span> <span>{priceData.vat?.toFixed(2)} {priceData.currency}</span></div> <div style={totalStyle}><span>Total:</span> <span>{priceData.total?.toFixed(2)} {priceData.currency}</span></div> </div> ); };
const RentalInfoDisplay = ({ rentalData }) => { if (!rentalData) return null; return ( <div style={{ marginBottom: '15px' }}> <h4 style={{ marginBottom: '10px' }}>Rental Information</h4> <p><strong>Pick-up:</strong> {rentalData.pickup_location || 'N/A'} <br /> {new Date(rentalData.pickup_datetime).toLocaleString()}</p> <p style={{marginTop: '5px'}}><strong>Return:</strong> {rentalData.return_location || rentalData.pickup_location || 'N/A'} <br /> {new Date(rentalData.dropoff_datetime).toLocaleString()}</p> </div> ); };
const CarSummaryDisplay = ({ carData }) => { if (!carData) return null; return ( <div style={{ textAlign: 'center', marginBottom: '15px' }}> <h4 style={{ marginBottom: '10px' }}>Car Rental Summary</h4> {carData.image_url && <img src={carData.image_url} alt={`${carData.brand} ${carData.model}`} style={{maxWidth: '150px', height: 'auto', marginBottom: '10px'}} />} <p style={{fontWeight: 'bold'}}>{carData.brand} {carData.model}</p> <p style={{fontSize: '0.9em'}}>Size: {carData.size || 'N/A'}</p> </div> ); };
// -------------------------------------------------------------------


const CheckoutSummary = () => {
  const { rentalId } = useParams();
  const navigate = useNavigate();

  const { data: rental, isLoading: isLoadingRental, isError: isErrorRental, error: errorRental } = useQuery({
    queryKey: ['rentalDetails', rentalId], queryFn: () => fetchRentalDetails(rentalId), enabled: !!rentalId, staleTime: Infinity,
  });
  const { data: price, isLoading: isLoadingPrice, isError: isErrorPrice, error: errorPrice } = useQuery({
    queryKey: ['rentalPrice', rentalId], queryFn: () => fetchRentalPriceDetails(rentalId), enabled: !!rentalId,
  });

  const handleNext = () => navigate(`/checkout/${rentalId}/user-info`);
  const handleBack = () => navigate(-1); // Go back to car selection

  const isLoading = isLoadingRental || isLoadingPrice;
  const queryError = errorRental || errorPrice;

  return (
    <div style={{ maxWidth: '800px', margin: '20px auto', padding: '20px', border: '1px solid #eee', borderRadius: '10px', backgroundColor: '#fff' }}>
      <h2 style={{ textAlign: 'center', marginBottom: '20px' }}>Booking Summary & Price</h2>
       <div style={{textAlign: 'center', marginBottom: '20px', color: 'gray'}}>Step 1 &gt; <strong style={{color: 'black'}}>Step 2 (Summary)</strong> &gt; Step 3 &gt; Step 4</div>

      {isLoading && <LoadingSpinner />}
      <ErrorMessage message={queryError?.message} />

      {!isLoading && !queryError && rental && (
        <div style={{ display: 'flex', gap: '30px', flexWrap: 'wrap-reverse' }}>
           <div style={{ flex: '1 1 300px' }}>
               <RentalInfoDisplay rentalData={rental} />
               {rental.car && ( <> <h4>Car Details</h4> <p>Type: {rental.car.type || 'N/A'}, Engine: {rental.car.engine || 'N/A'}</p> <p>Seats: {rental.car.seat || 'N/A'}, Doors: {rental.car.door || 'N/A'}</p> <p>Transmission: {rental.car.transmission || 'N/A'}, Luggage: {rental.car.luggage || 'N/A'}</p> </> )}
           </div>
            <div style={{ flex: '1 1 250px', borderLeft: '1px solid #eee', paddingLeft: '30px' }}>
               <CarSummaryDisplay carData={rental.car} />
               <PriceDetailsDisplay priceData={price} isLoading={isLoadingPrice} isError={isErrorPrice} error={errorPrice} />
            </div>
        </div>
      )}

       <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '30px', borderTop: '1px solid #eee', paddingTop: '20px' }}>
           <button onClick={handleBack} style={{ padding: '10px 20px' }}>Back</button>
           <button onClick={handleNext} disabled={isLoading || !!queryError || !rental || !price} style={{ padding: '10px 20px', backgroundColor: '#007bff', color: 'white', border: 'none', borderRadius: '5px'}}>Next: Enter Info</button>
       </div>
    </div>
  );
};
export default CheckoutSummary;