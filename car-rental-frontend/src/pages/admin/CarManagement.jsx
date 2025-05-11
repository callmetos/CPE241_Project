import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchAllCars, deleteCar, fetchBranches } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import CarForm from '../../components/admin/CarForm.jsx';
import '../../components/admin/AdminCommon.css';

const SortIcon = ({ direction }) => {
    if (!direction) return null;
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const CarManagement = () => {
    const queryClient = useQueryClient();
    const [showForm, setShowForm] = useState(false);
    const [editingCar, setEditingCar] = useState(null);
    const [error, setError] = useState('');

    const [uiFilters, setUiFilters] = useState({
        brand: '', model: '', branch_id: '', availability: '',
    });
    const [sortConfig, setSortConfig] = useState({ key: 'id', direction: 'ASC' });
    const [currentPage, setCurrentPage] = useState(1);
    const [itemsPerPage, setItemsPerPage] = useState(15);

    const { data: branches = [], isLoading: isLoadingBranches } = useQuery({
        queryKey: ['branches'],
        queryFn: fetchBranches,
        staleTime: Infinity,
    });

    const activeQueryParams = useMemo(() => ({
        ...uiFilters,
        page: currentPage,
        limit: itemsPerPage,
        sort_by: sortConfig.key,
        sort_dir: sortConfig.direction.toUpperCase(),
    }), [uiFilters, currentPage, itemsPerPage, sortConfig]);


    const {
        data: carsData,
        isLoading: isLoadingCars,
        isError: isErrorCars,
        error: carsError,
    } = useQuery({
        queryKey: ['cars', activeQueryParams],
        queryFn: () => fetchAllCars(activeQueryParams),
        staleTime: 1000 * 60 * 2,
        keepPreviousData: true,
    });

    const currentItemsToDisplay = carsData?.cars || [];
    const totalItems = carsData?.total_count || 0;
    const totalPages = Math.ceil(totalItems / itemsPerPage);


    const requestSort = useCallback((key) => {
        let direction = 'ASC';
        if (sortConfig.key === key && sortConfig.direction === 'ASC') {
            direction = 'DESC';
        }
        setSortConfig({ key, direction });
        setCurrentPage(1);
    }, [sortConfig]);

    const { mutate: removeCar, isPending: isDeleting } = useMutation({
        mutationFn: deleteCar,
        onSuccess: (data, carId) => {
            setError('');
            queryClient.invalidateQueries({ queryKey: ['cars'] });
            alert(`Car ID ${carId} deleted successfully!`);
        },
        onError: (err, carId) => {
            setError(`Failed to delete car ${carId}: ${err.message}`);
        },
    });

    const handleFilterChange = (e) => {
        const { name, value } = e.target;
        setUiFilters((prev) => ({ ...prev, [name]: value }));
        setCurrentPage(1);
    };

    const handleClearFilters = () => {
        setUiFilters({ brand: '', model: '', branch_id: '', availability: '' });
        setCurrentPage(1);
    };

    const handleAddCar = () => { setEditingCar(null); setShowForm(true); setError(''); };
    const handleEditCar = (car) => { setEditingCar(car); setShowForm(true); setError(''); };
    const handleDeleteCar = (carId) => {
        if (window.confirm(`Delete car ID ${carId}? This might fail if it has active rentals.`)) {
            setError('');
            removeCar(carId);
        }
    };
    const handleCloseForm = () => { setShowForm(false); setEditingCar(null); };

    const paginate = (pageNumber) => {
        if (pageNumber > 0 && pageNumber <= totalPages) {
            setCurrentPage(pageNumber);
        }
    };

    const imgStyle = { maxWidth: '70px', height: 'auto', borderRadius: '3px', display: 'block' };
    const filterSectionStyle = { display: 'flex', flexWrap: 'wrap', gap: '15px', padding: '15px', marginBottom: '20px', backgroundColor: '#f8f9fa', borderRadius: '4px', border: '1px solid var(--admin-border-color)' };
    const filterGroupStyle = { display: 'flex', flexDirection: 'column', flex: '1 1 180px' };
    const filterLabelStyle = { marginBottom: '5px', fontSize: '0.85em', fontWeight: '500', color: 'var(--admin-text-medium)' };
    const filterInputStyle = { padding: '8px', fontSize: '0.9em', border: '1px solid #ccc', borderRadius: '4px' };
    const filterButtonStyle = { alignSelf: 'flex-end', padding: '8px 15px' };
    const thSortableStyle = { cursor: 'pointer', userSelect: 'none' };

    const paginationContainerStyle = { display: 'flex', justifyContent: 'center', alignItems: 'center', marginTop: '20px', paddingTop: '15px', borderTop: '1px solid #eee' };
    const paginationButtonStyle = (isActive) => ({ margin: '0 5px', padding: '8px 12px', cursor: 'pointer', backgroundColor: isActive ? 'var(--admin-primary)' : '#f0f0f0', color: isActive ? 'white' : '#333', border: `1px solid ${isActive ? 'var(--admin-primary)' : '#ccc'}`, borderRadius: '4px', fontWeight: isActive ? 'bold' : 'normal', });
    const paginationNavButtonStyle = { ...paginationButtonStyle(false), backgroundColor: '#e9ecef' }

    return (
        <div className="admin-container">
            <div className="admin-header">
                <h2>Car Management</h2>
                <button onClick={handleAddCar} className="admin-button admin-button-primary">
                    + Add Car
                </button>
            </div>

            <div style={filterSectionStyle}>
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-brand" style={filterLabelStyle}>Brand</label>
                    <input type="text" id="filter-brand" name="brand" style={filterInputStyle} value={uiFilters.brand} onChange={handleFilterChange} placeholder="e.g., Toyota"/>
                </div>
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-model" style={filterLabelStyle}>Model</label>
                    <input type="text" id="filter-model" name="model" style={filterInputStyle} value={uiFilters.model} onChange={handleFilterChange} placeholder="e.g., Camry" />
                </div>
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-branch" style={filterLabelStyle}>Branch</label>
                    <select id="filter-branch" name="branch_id" style={filterInputStyle} value={uiFilters.branch_id} onChange={handleFilterChange} disabled={isLoadingBranches} >
                        <option value="">All Branches</option>
                        {isLoadingBranches ? (<option disabled>Loading...</option>) : (branches.map(branch => (<option key={branch.id} value={branch.id}>{branch.name}</option>)))}
                    </select>
                </div>
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-availability" style={filterLabelStyle}>Availability</label>
                    <select id="filter-availability" name="availability" style={filterInputStyle} value={uiFilters.availability} onChange={handleFilterChange} >
                        <option value="">All</option>
                        <option value="true">Available</option>
                        <option value="false">Unavailable</option>
                    </select>
                </div>
                <div style={{...filterGroupStyle, flexDirection: 'row', alignItems: 'flex-end', gap: '10px', flexBasis: 'auto' }}>
                    <button onClick={handleClearFilters} className="admin-button admin-button-secondary" style={filterButtonStyle} disabled={isLoadingCars}>Clear Filters</button>
                </div>
            </div>

            <ErrorMessage message={error || (isErrorCars ? `Error fetching cars: ${carsError?.message}` : null)} />
            {isLoadingCars && <LoadingSpinner />}

            {!isLoadingCars && (
            <>
                <div className="admin-table-wrapper">
                    <table className="admin-table">
                        <thead>
                            <tr>
                                <th style={thSortableStyle} onClick={() => requestSort('id')}>ID <SortIcon direction={sortConfig.key === 'id' ? sortConfig.direction.toLowerCase() : null} /></th>
                                <th>Image</th>
                                <th style={thSortableStyle} onClick={() => requestSort('brand')}>Brand <SortIcon direction={sortConfig.key === 'brand' ? sortConfig.direction.toLowerCase() : null} /></th>
                                <th style={thSortableStyle} onClick={() => requestSort('model')}>Model <SortIcon direction={sortConfig.key === 'model' ? sortConfig.direction.toLowerCase() : null} /></th>
                                <th style={thSortableStyle} onClick={() => requestSort('price_per_day')}>Price/Day <SortIcon direction={sortConfig.key === 'price_per_day' ? sortConfig.direction.toLowerCase() : null} /></th>
                                <th style={thSortableStyle} onClick={() => requestSort('branch_id')}>Branch ID <SortIcon direction={sortConfig.key === 'branch_id' ? sortConfig.direction.toLowerCase() : null} /></th>
                                <th style={thSortableStyle} onClick={() => requestSort('availability')}>Available <SortIcon direction={sortConfig.key === 'availability' ? sortConfig.direction.toLowerCase() : null} /></th>
                                <th>Spot</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {currentItemsToDisplay.length === 0 ? (
                                <tr className="admin-table-placeholder"><td colSpan="9">No cars found matching the criteria.</td></tr>
                            ) : (
                                currentItemsToDisplay.map((car) => (
                                    <tr key={car.id}>
                                        <td>{car.id}</td>
                                        <td><img src={car.image_url || `https://placehold.co/70x50/eee/ccc?text=N/A`} alt={`${car.brand} ${car.model}`} style={imgStyle} onError={(e) => { e.target.onerror = null; e.target.src='https://placehold.co/70x50/eee/ccc?text=Err'; }} /></td>
                                        <td>{car.brand}</td>
                                        <td>{car.model}</td>
                                        <td>{car.price_per_day?.toFixed(2)}</td>
                                        <td>{car.branch_id}</td>
                                        <td><span className={car.availability ? 'status-available-yes' : 'status-available-no'}>{car.availability ? 'Yes' : 'No'}</span></td>
                                        <td>{car.parking_spot || '-'}</td>
                                        <td className="actions admin-action-buttons">
                                            <button onClick={() => handleEditCar(car)} className="admin-button admin-button-warning admin-button-sm" disabled={isDeleting}>Edit</button>
                                            <button onClick={() => handleDeleteCar(car.id)} className="admin-button admin-button-danger admin-button-sm" disabled={isDeleting}>{isDeleting ? '...' : 'Delete'}</button>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
                 {totalPages > 1 && (
                    <div style={paginationContainerStyle}>
                         <button onClick={() => paginate(currentPage - 1)} disabled={currentPage === 1} style={paginationNavButtonStyle}>&laquo; Previous</button>
                        {Array.from({ length: totalPages }, (_, i) => {
                             const pageNumber = i + 1;
                             const showPage = pageNumber === 1 || pageNumber === totalPages || (pageNumber >= currentPage - 2 && pageNumber <= currentPage + 2);
                             let showEllipsisBefore = false;
                             let showEllipsisAfter = false;

                             if (totalPages > 5) {
                                if (currentPage > 3 && pageNumber === 2) showEllipsisBefore = true;
                                if (currentPage < totalPages - 2 && pageNumber === totalPages - 1) showEllipsisAfter = true;
                             }


                             if (showEllipsisBefore && pageNumber === 2 && !(pageNumber >= currentPage -2 && pageNumber <= currentPage + 2) ) return <span key={`ellipsis-start-${pageNumber}`} style={{ margin: '0 5px' }}>...</span>;
                             if (showEllipsisAfter && pageNumber === totalPages -1 && !(pageNumber >= currentPage -2 && pageNumber <= currentPage + 2) ) return <span key={`ellipsis-end-${pageNumber}`} style={{ margin: '0 5px' }}>...</span>;


                             if(showPage) {
                                return (<button key={pageNumber} onClick={() => paginate(pageNumber)} style={paginationButtonStyle(currentPage === pageNumber)}>{pageNumber}</button>);
                             }
                             return null;
                        })}
                        <button onClick={() => paginate(currentPage + 1)} disabled={currentPage === totalPages || totalPages === 0} style={paginationNavButtonStyle}>Next &raquo;</button>
                    </div>
                )}
            </>
            )}
            {showForm && (<CarForm initialData={editingCar} onClose={handleCloseForm} />)}
        </div>
    );
};

export default CarManagement;
