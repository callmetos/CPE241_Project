// src copy/pages/CheckoutPaymentUpload.jsx
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchRentalDetails, fetchRentalPriceDetails, uploadPaymentSlip } from '../services/apiService.js';
import LoadingSpinner from '../components/LoadingSpinner.jsx';
import ErrorMessage from '../components/ErrorMessage.jsx';

const BANK_ACCOUNT = {
    name: "Channathat Ueanapaphon",
    number: "123-4-56789-0",
    bank: "ธนาคารไทยพาณิชย์ (SCB)",
};

const PriceDetailsDisplay = ({ priceData }) => {
    if (!priceData) return <p>Loading price...</p>;
    return (
        <div style={{border:'1px dashed #eee', padding: '10px', margin:'10px 0', borderRadius: '4px', background: '#f8f9fa'}}>
            <p><strong>ยอดชำระ: {priceData.amount?.toFixed(2)} {priceData.currency || 'THB'}</strong></p>
        </div>
    );
};

const RentalInfoDisplay = ({ rentalData }) => rentalData ? <div style={{fontSize: '0.9em', color: '#555'}}><p><strong>การเช่า ID:</strong> {rentalData.id}</p><p><strong>รับรถ:</strong> {new Date(rentalData.pickup_datetime).toLocaleString()}</p><p><strong>คืนรถ:</strong> {new Date(rentalData.dropoff_datetime).toLocaleString()}</p></div> : null;
const CarSummaryDisplay = ({ carData }) => carData ? <div style={{textAlign: 'center'}}><p style={{fontWeight: 'bold'}}>{carData.brand} {carData.model}</p>{carData.image_url && <img src={carData.image_url} alt={`${carData.brand} ${carData.model}`} style={{maxWidth: '120px', height: 'auto', marginTop: '5px', borderRadius: '4px'}} onError={(e) => { e.target.onerror = null; e.target.src='https://placehold.co/120x80/eee/ccc?text=Car'; }} />}</div> : null;


const CheckoutPaymentUpload = () => {
    const { rentalId } = useParams();
    const navigate = useNavigate();
    const queryClient = useQueryClient();
    const [selectedFile, setSelectedFile] = useState(null);
    const [previewUrl, setPreviewUrl] = useState(null);
    const [uploadError, setUploadError] = useState('');
    const [generalError, setGeneralError] = useState('');

    const { data: rental, isLoading: isLoadingRental, isError: isErrorRental, error: errorRental } = useQuery({
        queryKey: ['rentalDetails', rentalId],
        queryFn: () => fetchRentalDetails(rentalId),
        enabled: !!rentalId,
        staleTime: 1000 * 60 * 5,
    });

    const { data: price, isLoading: isLoadingPrice, isError: isErrorPrice, error: errorPrice } = useQuery({
        queryKey: ['rentalPrice', rentalId],
        queryFn: () => fetchRentalPriceDetails(rentalId),
        enabled: !!rentalId,
        staleTime: 1000 * 60 * 5,
        onSuccess: (data) => {
            console.log("Fetched Price Details in CheckoutPaymentUpload:", data);
        }
    });

    const { mutate: submitSlip, isPending: isUploading } = useMutation({
        mutationFn: ({ rentalId, file }) => uploadPaymentSlip(rentalId, file),
        onSuccess: (data) => {
            alert('อัปโหลดสลิปสำเร็จแล้ว โปรดรอการตรวจสอบ');
            navigate('/rental-history', { replace: true });
            queryClient.invalidateQueries({ queryKey: ['rentalDetails', rentalId] });
            queryClient.invalidateQueries({ queryKey: ['rentalHistory'] });
        },
        onError: (error) => {
            setUploadError(`อัปโหลดสลิปไม่สำเร็จ: ${error.message}`);
        },
    });

    useEffect(() => {
        if (!selectedFile) {
            setPreviewUrl(null);
            return;
        }
        const objectUrl = URL.createObjectURL(selectedFile);
        setPreviewUrl(objectUrl);
        return () => URL.revokeObjectURL(objectUrl);
    }, [selectedFile]);

    const handleFileChange = (event) => {
        const file = event.target.files?.[0];
        if (file) {
            if (!file.type.startsWith('image/')) {
                setUploadError('กรุณาเลือกไฟล์รูปภาพเท่านั้น (JPG, PNG, GIF)');
                setSelectedFile(null);
                event.target.value = '';
                return;
            }
            setSelectedFile(file);
            setUploadError('');
        } else {
            setSelectedFile(null);
        }
    };

    const handleUpload = () => {
        if (!selectedFile) {
            setUploadError('กรุณาเลือกไฟล์สลิปก่อน');
            return;
        }
        if (!rentalId) {
             setUploadError('ไม่พบข้อมูลการเช่า');
             return;
        }
        setUploadError('');
        submitSlip({ rentalId, file: selectedFile });
    };

    const handleBack = () => {
        navigate(`/checkout/${rentalId}/user-info`);
    };

    const isLoading = isLoadingRental || isLoadingPrice;
    const queryError = errorRental || errorPrice;

    useEffect(() => {
        if (queryError) {
            setGeneralError(`เกิดข้อผิดพลาดในการโหลดข้อมูล: ${queryError.message}`);
        } else {
            setGeneralError('');
        }
    }, [queryError]);

    const containerStyle = { maxWidth: '800px', margin: '20px auto', padding: '30px', border: '1px solid #e0e0e0', borderRadius: '8px', backgroundColor: '#fff', boxShadow: '0 3px 8px rgba(0,0,0,0.05)' };
    const headerStyle = { textAlign: 'center', marginBottom: '25px', color: '#333' };
    const sectionStyle = { marginBottom: '25px', paddingBottom: '20px', borderBottom: '1px solid #eee' };
    const bankDetailsStyle = { backgroundColor: '#f8f9fa', padding: '15px', borderRadius: '5px', border: '1px solid #eee', marginBottom: '15px' };
    const uploadAreaStyle = { border: '2px dashed #ccc', padding: '20px', textAlign: 'center', borderRadius: '5px', backgroundColor: '#fafafa', cursor: 'pointer', transition: 'border-color 0.2s' };
    const fileInputStyle = { display: 'none' };
    const previewImageStyle = { maxWidth: '200px', maxHeight: '200px', marginTop: '15px', borderRadius: '4px', border: '1px solid #ddd' };
    const buttonContainerStyle = { display: 'flex', justifyContent: 'space-between', marginTop: '30px', paddingTop: '20px', borderTop: '1px solid #eee' };
    const buttonStyle = { padding: '10px 20px', borderRadius: '5px', cursor: 'pointer', border: 'none', fontSize: '1rem', fontWeight: '500' };
    const backButtonStyle = { ...buttonStyle, backgroundColor: '#6c757d', color: 'white' };
    const uploadButtonStyle = { ...buttonStyle, backgroundColor: '#198754', color: 'white' };
    const disabledButtonStyle = { ...uploadButtonStyle, backgroundColor: '#cccccc', cursor: 'not-allowed', opacity: 0.7 };

    return (
        <div style={containerStyle}>
            <h2 style={headerStyle}>ชำระเงินและอัปโหลดสลิป</h2>
            {isLoading && <LoadingSpinner />}
            <ErrorMessage message={generalError} />

            {!isLoading && !queryError && rental && price && (
                <div>
                    <div style={sectionStyle}>
                        <h4 style={{ marginBottom: '15px' }}>ข้อมูลการชำระเงิน</h4>
                        <div style={{ display: 'flex', gap: '20px', flexWrap: 'wrap' }}>
                            <div style={{ flex: '1 1 300px' }}>
                                <p>กรุณาโอนเงินมาที่บัญชี:</p>
                                <div style={bankDetailsStyle}>
                                    <p><strong>ชื่อบัญชี:</strong> {BANK_ACCOUNT.name}</p>
                                    <p><strong>เลขที่บัญชี:</strong> {BANK_ACCOUNT.number}</p>
                                    <p><strong>ธนาคาร:</strong> {BANK_ACCOUNT.bank}</p>
                                </div>
                                <PriceDetailsDisplay priceData={price} />
                                <p style={{ fontSize: '0.9em', color: 'red', marginTop: '10px' }}>* โปรดชำระเงินภายใน 24 ชั่วโมงและอัปโหลดสลิปเพื่อยืนยันการจอง</p>
                            </div>
                            <div style={{ flex: '1 1 200px', borderLeft: '1px solid #eee', paddingLeft: '20px' }}>
                                <h5 style={{marginBottom: '10px'}}>สรุปการจอง</h5>
                                <CarSummaryDisplay carData={rental?.car} />
                                <hr style={{margin: '10px 0', borderColor: '#eee'}}/>
                                <RentalInfoDisplay rentalData={rental} />
                            </div>
                        </div>
                    </div>
                    <div>
                        <h4 style={{ marginBottom: '15px' }}>อัปโหลดสลิปโอนเงิน</h4>
                        <label htmlFor="slipUpload" style={uploadAreaStyle}>
                            {previewUrl ? 'คลิกเพื่อเปลี่ยนรูปภาพ' : 'คลิก หรือ ลากไฟล์สลิปมาวางที่นี่'}
                        </label>
                        <input
                            id="slipUpload"
                            type="file"
                            accept="image/png, image/jpeg, image/gif"
                            onChange={handleFileChange}
                            style={fileInputStyle}
                        />
                        {previewUrl && (
                            <div style={{ textAlign: 'center', marginTop: '15px' }}>
                                <p>รูปภาพที่เลือก:</p>
                                <img src={previewUrl} alt="Payment Slip Preview" style={previewImageStyle} />
                            </div>
                        )}
                        <ErrorMessage message={uploadError} />
                    </div>
                    <div style={buttonContainerStyle}>
                        <button onClick={handleBack} style={backButtonStyle} disabled={isUploading}>
                            กลับ
                        </button>
                        <button
                            onClick={handleUpload}
                            style={isUploading ? disabledButtonStyle : uploadButtonStyle}
                            disabled={!selectedFile || isUploading}
                        >
                            {isUploading ? 'กำลังอัปโหลด...' : 'ยืนยันและอัปโหลดสลิป'}
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default CheckoutPaymentUpload;