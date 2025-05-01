import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchRevenueReport, fetchPopularCarsReport, fetchBranchPerformanceReport } from '../../services/apiService';
import LoadingSpinner from '../../components/LoadingSpinner';
import ErrorMessage from '../../components/ErrorMessage';
import '../../components/admin/AdminCommon.css';



const RevenueReport = () => {
    const today = new Date().toISOString().split('T')[0];
    const lastWeek = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const [startDate, setStartDate] = useState(lastWeek);
    const [endDate, setEndDate] = useState(today);
    const [fetchEnabled, setFetchEnabled] = useState(true);

    const { data: reportData = [], isLoading, isError, error, refetch } = useQuery({
        queryKey: ['revenueReport', startDate, endDate],
        queryFn: () => fetchRevenueReport(startDate, endDate),
        enabled: fetchEnabled,
        staleTime: 1000 * 60 * 5,
    });

    const handleFetch = () => {
        if (startDate && endDate && new Date(endDate) >= new Date(startDate)) {
             setFetchEnabled(true);
             refetch();
        } else {
             alert("Please select valid start and end dates.");
             setFetchEnabled(false);
        }
    }

    return (
        <div className="admin-card">
            <h3 className="admin-card-title">Revenue Report (by Day)</h3>
            <div style={{ display: 'flex', gap: '15px', alignItems: 'center', margin: '15px 0', flexWrap: 'wrap' }}>
                <div>
                    <label htmlFor="startDate" style={{ marginRight: '5px', fontSize: '0.9em' }}>Start:</label>
                    <input type="date" id="startDate" value={startDate} onChange={(e) => { setStartDate(e.target.value); setFetchEnabled(false); }} style={{ padding: '5px', border: '1px solid #ccc' }} />
                </div>
                <div>
                    <label htmlFor="endDate" style={{ marginRight: '5px', fontSize: '0.9em' }}>End:</label>
                    <input type="date" id="endDate" value={endDate} onChange={(e) => { setEndDate(e.target.value); setFetchEnabled(false); }} style={{ padding: '5px', border: '1px solid #ccc' }} />
                </div>
                <button onClick={handleFetch} className="admin-button admin-button-info admin-button-sm" disabled={isLoading}>
                    {isLoading ? 'Loading...' : 'Fetch Report'}
                </button>
            </div>
            <ErrorMessage message={isError ? `Error: ${error?.message}` : null} />
            {isLoading ? <LoadingSpinner /> : (
                <div className="admin-table-wrapper">
                    <table className="admin-table">
                        <thead>
                            <tr><th>Date</th><th>Total Revenue (THB)</th></tr>
                        </thead>
                        <tbody>
                            {reportData.length === 0 && !isError && <tr><td colSpan="2">No revenue data found for this period.</td></tr>}
                            {reportData.map(item => (
                                <tr key={item.period}>
                                    <td>{item.period}</td>
                                    <td style={{ textAlign: 'right' }}>{item.amount.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
};

const PopularCarsReport = () => {
    const [limit, setLimit] = useState(10);
    const { data: reportData = [], isLoading, isError, error } = useQuery({
        queryKey: ['popularCarsReport', limit],
        queryFn: () => fetchPopularCarsReport(limit),
        staleTime: 1000 * 60 * 10,
    });

     const handleLimitChange = (e) => {
        const newLimit = parseInt(e.target.value, 10);
        if (newLimit > 0) {
            setLimit(newLimit);
        }
     }

    return (
        <div className="admin-card">
            <h3 className="admin-card-title">Most Popular Cars</h3>
             <div style={{margin: '10px 0'}}>
                <label htmlFor="popLimit" style={{ marginRight: '5px', fontSize: '0.9em' }}>Show Top:</label>
                <select id="popLimit" value={limit} onChange={handleLimitChange} style={{ padding: '5px', border: '1px solid #ccc' }}>
                    <option value="5">5</option>
                    <option value="10">10</option>
                    <option value="20">20</option>
                </select>
             </div>
            <ErrorMessage message={isError ? `Error: ${error?.message}` : null} />
            {isLoading ? <LoadingSpinner /> : (
                <div className="admin-table-wrapper">
                    <table className="admin-table">
                        <thead>
                            <tr><th>Rank</th><th>Car (Brand Model)</th><th>Rental Count</th></tr>
                        </thead>
                        <tbody>
                             {reportData.length === 0 && !isError && <tr><td colSpan="3">No rental data found.</td></tr>}
                            {reportData.map((item, index) => (
                                <tr key={item.car_id}>
                                    <td>{index + 1}</td>
                                    <td>{item.brand} {item.model} (ID: {item.car_id})</td>
                                    <td>{item.rental_count}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
};

const BranchPerformanceReport = () => {
    const { data: reportData = [], isLoading, isError, error } = useQuery({
        queryKey: ['branchPerformanceReport'],
        queryFn: fetchBranchPerformanceReport,
        staleTime: 1000 * 60 * 10,
    });

    return (
        <div className="admin-card">
            <h3 className="admin-card-title">Branch Performance Summary</h3>
             <ErrorMessage message={isError ? `Error: ${error?.message}` : null} />
            {isLoading ? <LoadingSpinner /> : (
                 <div className="admin-table-wrapper">
                    <table className="admin-table">
                        <thead>
                            <tr><th>Branch</th><th>Total Rentals</th><th>Total Revenue (THB)</th></tr>
                        </thead>
                        <tbody>
                             {reportData.length === 0 && !isError && <tr><td colSpan="3">No branch data found.</td></tr>}
                            {reportData.map(item => (
                                <tr key={item.branch_id}>
                                    <td>{item.branch_name} (ID: {item.branch_id})</td>
                                    <td>{item.total_rentals}</td>
                                     <td style={{ textAlign: 'right' }}>{item.total_revenue.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                 </div>
            )}
        </div>
    );
};



const Reports = () => {
    return (
        <div className="admin-container">
            <div className="admin-header">
                <h2>Reports</h2>
            </div>

            <div style={{ display: 'grid', gap: '25px' }}>
                <RevenueReport />
                <PopularCarsReport />
                <BranchPerformanceReport />
            </div>
        </div>
    );
};

export default Reports;