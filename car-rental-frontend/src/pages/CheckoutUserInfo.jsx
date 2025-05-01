import React, { useState, useContext, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { AuthContext } from '../context/AuthContext.jsx';
import { fetchRentalDetails, fetchRentalPriceDetails } from '../services/apiService.js';
import LoadingSpinner from '../components/LoadingSpinner.jsx';
import ErrorMessage from '../components/ErrorMessage.jsx';
// import CheckoutStepper from '../components/CheckoutStepper.jsx'; // Optional
import '../components/admin/AdminForm.css'; // ใช้สไตล์ฟอร์มร่วมกัน
import '../components/admin/AdminCommon.css'; // ใช้สไตล์ปุ่มร่วมกัน

// --- Reusable Display Components (Placeholders - ใช้ Component จริงของคุณ) ---
const PriceDetailsDisplay = ({ priceData }) => priceData ? <div style={{border:'1px dashed #eee', padding: '10px', margin:'10px 0', borderRadius: '4px', background: '#f8f9fa'}}><p><strong>Total: {priceData.total?.toFixed(2)} {priceData.currency}</strong></p></div> : <p>Loading price...</p>;
const RentalInfoDisplay = ({ rentalData }) => rentalData ? <div style={{fontSize: '0.9em', color: '#555'}}><p><strong>Rental ID:</strong> {rentalData.id}</p><p><strong>Pickup:</strong> {new Date(rentalData.pickup_datetime).toLocaleString()}</p><p><strong>Return:</strong> {new Date(rentalData.dropoff_datetime).toLocaleString()}</p></div> : null;
const CarSummaryDisplay = ({ carData }) => carData ? <div style={{textAlign: 'center'}}><p style={{fontWeight: 'bold'}}>{carData.brand} {carData.model}</p>{carData.image_url && <img src={carData.image_url} alt={`${carData.brand} ${carData.model}`} style={{maxWidth: '120px', height: 'auto', marginTop: '5px', borderRadius: '4px'}} onError={(e) => { e.target.onerror = null; e.target.src='https://placehold.co/120x80/eee/ccc?text=Car'; }} />}</div> : null;
// ----------------------------------------------------

const CheckoutUserInfo = () => {
    const { rentalId } = useParams(); // *** รับ rentalId จาก URL ***
    const navigate = useNavigate();
    const { user: loggedInUser } = useContext(AuthContext);

    // State สำหรับข้อมูลในฟอร์ม
    const [formData, setFormData] = useState({
        email: '',
        firstName: '',
        lastName: '',
        phone: '',
        drivingLicense: ''
    });
    const [formError, setFormError] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);

    // Fetch Rental and Price data (อาจจะดึงแค่บางส่วน)
    const { data: rental, isLoading: isLoadingRental, isError: isErrorRental, error: errorRental } = useQuery({
        queryKey: ['rentalDetails', rentalId], queryFn: () => fetchRentalDetails(rentalId), enabled: !!rentalId, staleTime: 1000 * 60 * 5
    });
    const { data: price, isLoading: isLoadingPrice, isError: isErrorPrice, error: errorPrice } = useQuery({
        queryKey: ['rentalPrice', rentalId], queryFn: () => fetchRentalPriceDetails(rentalId), enabled: !!rentalId && !!rental, staleTime: 1000 * 60 * 5
    });

    // Pre-fill form with logged-in user data
    useEffect(() => {
        if (loggedInUser) {
            setFormData(prev => ({
                ...prev,
                email: loggedInUser.email || '',
                firstName: loggedInUser.name?.split(' ')[0] || '',
                lastName: loggedInUser.name?.split(' ').slice(1).join(' ') || '',
                phone: loggedInUser.phone || ''
            }));
        }
    }, [loggedInUser]);

    // --- Event Handlers ---
    const handleInputChange = (e) => {
        setFormData(prev => ({ ...prev, [e.target.name]: e.target.value }));
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        setFormError('');
        setIsSubmitting(true);

        // Validation
        if (!formData.firstName || !formData.lastName || !formData.phone || !formData.drivingLicense || !formData.email) {
            setFormError("กรุณากรอกข้อมูลที่จำเป็น (*) ให้ครบถ้วน");
            setIsSubmitting(false);
            return;
        }
        // TODO: เพิ่ม Validation อื่นๆ ถ้าต้องการ

        // --- Navigate ไปหน้า Payment Upload ---
        console.log("User info submitted, navigating to payment upload for rental:", rentalId);
        // Backend ควรจะบันทึกข้อมูลผู้ใช้เหล่านี้เมื่อมีการยืนยัน Payment หรือสร้าง Payment record
        // เราไม่จำเป็นต้องส่งข้อมูลนี้ไปใน state ถ้า Backend จัดการได้
        navigate(`/checkout/${rentalId}/payment-upload`);

        // ไม่ต้อง setIsSubmitting(false) เพราะเปลี่ยนหน้า
    };

    const handleBack = () => {
        // กลับไปหน้า Summary
        navigate(`/checkout/${rentalId}/summary`);
    };

    // --- Loading and Error States ---
    const isLoading = isLoadingRental || isLoadingPrice;
    const queryError = errorRental || errorPrice;

    return (
        <div className="admin-container" style={{ maxWidth: '950px' }}>
            {/* <CheckoutStepper currentStep={3} /> */}
            <h2 style={{ textAlign: 'center', marginBottom: '25px', color: '#333' }}>Enter Renter Information</h2>

            {isLoading && <LoadingSpinner />}
            <ErrorMessage message={queryError?.message || formError} />

            {!isLoading && !queryError && rental && price && (
                <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '30px' }}>
                    {/* Form Column */}
                    <div>
                        {/* ใช้ class จาก AdminForm.css */}
                        <form onSubmit={handleSubmit} className="admin-form-container" style={{ maxWidth: '100%', boxShadow: 'none', border: 'none', padding: 0 }}>
                            <h4 style={{ marginBottom: '15px', fontWeight: '500' }}>Your Information</h4>
                            <p style={{ fontSize: '0.85em', color: 'grey', marginBottom: '20px' }}>Please fill in the information in English.</p>

                            <div className="admin-form-group">
                                <label htmlFor="email" className="admin-form-label">Email *</label>
                                <input type="email" id="email" name="email" value={formData.email} onChange={handleInputChange} className="admin-form-input" required readOnly={!!loggedInUser?.email} disabled={isSubmitting} />
                            </div>

                            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px' }}>
                                <div className="admin-form-group">
                                    <label htmlFor="firstName" className="admin-form-label">First Name *</label>
                                    <input type="text" id="firstName" name="firstName" value={formData.firstName} onChange={handleInputChange} className="admin-form-input" required disabled={isSubmitting} />
                                </div>
                                <div className="admin-form-group">
                                    <label htmlFor="lastName" className="admin-form-label">Last Name *</label>
                                    <input type="text" id="lastName" name="lastName" value={formData.lastName} onChange={handleInputChange} className="admin-form-input" required disabled={isSubmitting} />
                                </div>
                            </div>

                            <div className="admin-form-group">
                                <label htmlFor="phone" className="admin-form-label">Phone *</label>
                                <input type="tel" id="phone" name="phone" value={formData.phone} onChange={handleInputChange} className="admin-form-input" required disabled={isSubmitting} />
                            </div>

                            <div className="admin-form-group">
                                <label htmlFor="drivingLicense" className="admin-form-label">Driving License Number *</label>
                                <input type="text" id="drivingLicense" name="drivingLicense" value={formData.drivingLicense} onChange={handleInputChange} className="admin-form-input" required disabled={isSubmitting} />
                            </div>

                            {/* Action Buttons - ใช้ class จาก AdminCommon.css */}
                            <div className="admin-form-button-container" style={{ borderTop: '1px solid #eee', marginTop: '30px', paddingTop: '20px' }}>
                                <button type="button" onClick={handleBack} className="admin-button admin-button-secondary" disabled={isSubmitting}>
                                    Back
                                </button>
                                <button type="submit" className="admin-button admin-button-primary" disabled={isSubmitting}>
                                    {isSubmitting ? 'Processing...' : 'Proceed to Payment'}
                                </button>
                            </div>
                        </form>
                    </div>

                    {/* Summary Column */}
                    <div style={{ borderLeft: '1px solid #eee', paddingLeft: '30px' }}>
                        <h4 style={{ marginBottom: '15px', fontWeight: '500' }}>Booking Summary</h4>
                        <CarSummaryDisplay carData={rental?.car} />
                        <hr style={{margin: '15px 0', borderColor: '#eee'}}/>
                        <RentalInfoDisplay rentalData={rental} />
                        <PriceDetailsDisplay priceData={price} />
                    </div>
                </div>
            )}
            {/* Show message if data couldn't be loaded */}
            {!isLoading && (!rental || !price) && <ErrorMessage message="Could not load booking details." />}
        </div>
    );
};
export default CheckoutUserInfo;
