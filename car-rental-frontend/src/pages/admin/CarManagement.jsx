import React, { useState, useMemo, useEffect, useCallback } from 'react'; // Add useCallback
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
// Import fetchBranches
import { fetchAllCars, deleteCar, fetchBranches } from '../../services/apiService.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import ErrorMessage from '../../components/ErrorMessage.jsx';
import CarForm from '../../components/admin/CarForm.jsx';
import '../../components/admin/AdminCommon.css'; // Import common styles

// --- Sort Icon Component (Optional) ---
const SortIcon = ({ direction }) => {
    if (!direction) return null; // No icon if unsorted
    // Simple text icons, consider using actual icons (e.g., from react-icons) for better UI
    return direction === 'ascending' ? ' ▲' : ' ▼';
};

const CarManagement = () => {
    const queryClient = useQueryClient();
    const [showForm, setShowForm] = useState(false);
    const [editingCar, setEditingCar] = useState(null);
    const [error, setError] = useState('');
    // --- State for filters ---
    const [filters, setFilters] = useState({
        brand: '', model: '', branch_id: '', availability: '',
    });
    // State to trigger refetch only when Apply Filters is clicked
    const [activeFilters, setActiveFilters] = useState({});
    // --- State for sorting ---
    const [sortConfig, setSortConfig] = useState({ key: 'id', direction: 'ascending' }); // Default sort by ID asc

    // --- Fetch Branches for filter dropdown ---
    const { data: branches = [], isLoading: isLoadingBranches } = useQuery({
        queryKey: ['branches'], // Query key for branches
        queryFn: fetchBranches, // Function to fetch branches
        staleTime: Infinity, // Cache branches indefinitely as they don't change often
    });

    // --- Fetch cars data using activeFilters ---
    // Note: Backend sorting would be more efficient for large datasets.
    // If implementing backend sort, pass sortConfig to fetchAllCars(activeFilters, sortConfig)
    const {
        data: cars = [], // Default to empty array
        isLoading: isLoadingCars,
        isError: isErrorCars,
        error: carsError,
        refetch: refetchCars, // Get refetch function from useQuery
    } = useQuery({
        // Query key includes activeFilters so it refetches when filters change
        queryKey: ['cars', activeFilters],
        // Pass activeFilters to the API call
        queryFn: () => fetchAllCars(activeFilters),
        staleTime: 1000 * 60 * 2, // Cache data for 2 minutes
        enabled: false, // Disable automatic fetching on mount; we'll trigger manually
    });

    // --- Trigger initial fetch and refetch when activeFilters change ---
    useEffect(() => {
        // console.log("Active filters changed, refetching cars:", activeFilters);
        refetchCars(); // Call the refetch function provided by useQuery
    }, [activeFilters, refetchCars]); // Dependency array includes activeFilters and refetchCars

    // --- Sorting Logic using useMemo ---
    const sortedCars = useMemo(() => {
        let sortableItems = [...cars]; // Create a mutable copy from the fetched data
        if (sortConfig.key !== null) {
            sortableItems.sort((a, b) => {
                // Handle potential null or undefined values safely
                const aValue = a[sortConfig.key] ?? ''; // Default to empty string or 0/false if needed
                const bValue = b[sortConfig.key] ?? '';

                // Determine comparison type based on the key
                if (sortConfig.key === 'price_per_day' || sortConfig.key === 'id' || sortConfig.key === 'branch_id') {
                    // Numeric comparison
                    const numA = parseFloat(aValue) || 0; // Handle non-numeric values gracefully
                    const numB = parseFloat(bValue) || 0;
                    if (numA < numB) return sortConfig.direction === 'ascending' ? -1 : 1;
                    if (numA > numB) return sortConfig.direction === 'ascending' ? 1 : -1;
                    return 0; // Numbers are equal
                } else if (sortConfig.key === 'availability') {
                    // Boolean comparison (treat true as "greater" than false)
                    const boolA = Boolean(aValue);
                    const boolB = Boolean(bValue);
                    if (boolA === boolB) return 0;
                    if (sortConfig.direction === 'ascending') {
                        return boolA ? 1 : -1; // false comes first when ascending
                    } else {
                        return boolA ? -1 : 1; // true comes first when descending
                    }
                } else {
                    // String comparison (case-insensitive)
                    if (String(aValue).toLowerCase() < String(bValue).toLowerCase()) {
                        return sortConfig.direction === 'ascending' ? -1 : 1;
                    }
                    if (String(aValue).toLowerCase() > String(bValue).toLowerCase()) {
                        return sortConfig.direction === 'ascending' ? 1 : -1;
                    }
                    return 0; // Strings are equal (case-insensitive)
                }
            });
        }
        return sortableItems;
    }, [cars, sortConfig]); // Recalculate when cars data or sortConfig changes

    // --- Request Sort Function ---
    // useCallback ensures this function is not recreated on every render unless sortConfig changes
    const requestSort = useCallback((key) => {
        let direction = 'ascending';
        // If sorting the same key, toggle direction
        if (sortConfig.key === key && sortConfig.direction === 'ascending') {
            direction = 'descending';
        }
        // If sorting a new key, default to ascending
        // Update the sort configuration state
        setSortConfig({ key, direction });
    }, [sortConfig]); // Dependency: only recreate if sortConfig changes

    // --- Mutation for deleting a car ---
    const { mutate: removeCar, isPending: isDeleting } = useMutation({
        mutationFn: deleteCar, // Function to call for deletion
        onSuccess: (data, carId) => {
            setError(''); // Clear any previous errors
            // Invalidate the 'cars' query with the current active filters to refetch the list
            queryClient.invalidateQueries({ queryKey: ['cars', activeFilters] });
            alert(`Car ID ${carId} deleted successfully!`);
        },
        onError: (err, carId) => {
            // Display error message if deletion fails
            setError(`Failed to delete car ${carId}: ${err.message}`);
        },
    });

    // --- Filter Handlers ---
    // Update the temporary filter state as the user types/selects
    const handleFilterChange = (e) => {
        const { name, value } = e.target;
        setFilters((prev) => ({ ...prev, [name]: value }));
    };

    // Apply the filters: update activeFilters to trigger the query refetch
    const handleApplyFilters = () => {
        // Create a clean filter object, removing empty values to avoid sending empty params
        const cleanFilters = {};
        Object.keys(filters).forEach(key => {
            if (filters[key] !== '' && filters[key] !== null && filters[key] !== undefined) {
                cleanFilters[key] = filters[key];
            }
        });
        setActiveFilters(cleanFilters); // Update active filters state
    };

    // Clear all filters and fetch all data
    const handleClearFilters = () => {
        setFilters({ brand: '', model: '', branch_id: '', availability: '' }); // Reset temporary filters
        setActiveFilters({}); // Reset active filters to empty object
    };

    // --- UI Handlers ---
    const handleAddCar = () => { setEditingCar(null); setShowForm(true); setError(''); };
    const handleEditCar = (car) => { setEditingCar(car); setShowForm(true); setError(''); };
    const handleDeleteCar = (carId) => {
        if (window.confirm(`Delete car ID ${carId}? This might fail if it has active rentals.`)) {
            setError('');
            removeCar(carId); // Trigger the delete mutation
        }
    };
    const handleCloseForm = () => { setShowForm(false); setEditingCar(null); };

    // --- Styles --- (Keep styles defined previously)
    const imgStyle = { maxWidth: '70px', height: 'auto', borderRadius: '3px', display: 'block' };
    const filterSectionStyle = { display: 'flex', flexWrap: 'wrap', gap: '15px', padding: '15px', marginBottom: '20px', backgroundColor: '#f8f9fa', borderRadius: '4px', border: '1px solid var(--admin-border-color)' };
    const filterGroupStyle = { display: 'flex', flexDirection: 'column', flex: '1 1 180px' };
    const filterLabelStyle = { marginBottom: '5px', fontSize: '0.85em', fontWeight: '500', color: 'var(--admin-text-medium)' };
    const filterInputStyle = { padding: '8px', fontSize: '0.9em', border: '1px solid #ccc', borderRadius: '4px' };
    const filterButtonStyle = { alignSelf: 'flex-end', padding: '8px 15px' };
    const thSortableStyle = { cursor: 'pointer', userSelect: 'none' }; // Style for clickable headers

    return (
        <div className="admin-container">
            {/* Header */}
            <div className="admin-header">
                <h2>Car Management</h2>
                <button onClick={handleAddCar} className="admin-button admin-button-primary">
                    + Add Car
                </button>
            </div>

            {/* Filter Section UI */}
            <div style={filterSectionStyle}>
                {/* Brand Filter */}
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-brand" style={filterLabelStyle}>Brand</label>
                    <input
                        type="text"
                        id="filter-brand"
                        name="brand"
                        style={filterInputStyle}
                        value={filters.brand}
                        onChange={handleFilterChange}
                        placeholder="e.g., Toyota"
                    />
                </div>
                {/* Model Filter */}
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-model" style={filterLabelStyle}>Model</label>
                    <input
                        type="text"
                        id="filter-model"
                        name="model"
                        style={filterInputStyle}
                        value={filters.model}
                        onChange={handleFilterChange}
                        placeholder="e.g., Camry"
                    />
                </div>
                {/* Branch Filter */}
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-branch" style={filterLabelStyle}>Branch</label>
                    <select
                        id="filter-branch"
                        name="branch_id"
                        style={filterInputStyle}
                        value={filters.branch_id}
                        onChange={handleFilterChange}
                        disabled={isLoadingBranches} // Disable while loading branches
                    >
                        <option value="">All Branches</option>
                        {isLoadingBranches ? (
                            <option disabled>Loading...</option>
                        ) : (
                            // Populate dropdown with fetched branches
                            branches.map(branch => (
                                <option key={branch.id} value={branch.id}>{branch.name}</option>
                            ))
                        )}
                    </select>
                </div>
                {/* Availability Filter */}
                <div style={filterGroupStyle}>
                    <label htmlFor="filter-availability" style={filterLabelStyle}>Availability</label>
                    <select
                        id="filter-availability"
                        name="availability"
                        style={filterInputStyle}
                        value={filters.availability}
                        onChange={handleFilterChange}
                    >
                        <option value="">All</option>
                        <option value="true">Available</option>
                        <option value="false">Unavailable</option>
                    </select>
                </div>
                {/* Filter Action Buttons */}
                <div style={{...filterGroupStyle, flexDirection: 'row', alignItems: 'flex-end', gap: '10px', flexBasis: 'auto' }}>
                    <button onClick={handleApplyFilters} className="admin-button admin-button-info" style={filterButtonStyle} disabled={isLoadingCars}>Apply Filters</button>
                    <button onClick={handleClearFilters} className="admin-button admin-button-secondary" style={filterButtonStyle} disabled={isLoadingCars}>Clear</button>
                </div>
            </div>

            {/* Display Error Messages */}
            <ErrorMessage message={error || (isErrorCars ? `Error fetching cars: ${carsError?.message}` : null)} />
            {/* Display Loading Spinner */}
            {isLoadingCars && <LoadingSpinner />}

            {/* Car Table */}
            {!isLoadingCars && ( // Render table only when not loading
                <div className="admin-table-wrapper">
                    <table className="admin-table">
                        <thead>
                            <tr>
                                {/* Sortable Table Headers */}
                                <th style={thSortableStyle} onClick={() => requestSort('id')}>
                                    ID <SortIcon direction={sortConfig.key === 'id' ? sortConfig.direction : null} />
                                </th>
                                <th>Image</th>
                                <th style={thSortableStyle} onClick={() => requestSort('brand')}>
                                    Brand <SortIcon direction={sortConfig.key === 'brand' ? sortConfig.direction : null} />
                                </th>
                                <th style={thSortableStyle} onClick={() => requestSort('model')}>
                                    Model <SortIcon direction={sortConfig.key === 'model' ? sortConfig.direction : null} />
                                </th>
                                <th style={thSortableStyle} onClick={() => requestSort('price_per_day')}>
                                    Price/Day <SortIcon direction={sortConfig.key === 'price_per_day' ? sortConfig.direction : null} />
                                </th>
                                <th style={thSortableStyle} onClick={() => requestSort('branch_id')}>
                                    Branch ID <SortIcon direction={sortConfig.key === 'branch_id' ? sortConfig.direction : null} />
                                </th>
                                <th style={thSortableStyle} onClick={() => requestSort('availability')}>
                                    Available <SortIcon direction={sortConfig.key === 'availability' ? sortConfig.direction : null} />
                                </th>
                                <th>Spot</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {/* Render sorted and filtered car data */}
                            {sortedCars.length === 0 ? (
                                <tr className="admin-table-placeholder">
                                    <td colSpan="9">No cars found matching the criteria.</td>
                                </tr>
                            ) : (
                                sortedCars.map((car) => (
                                    <tr key={car.id}>
                                        <td>{car.id}</td>
                                        <td>
                                            {/* Car Image with Fallback */}
                                            <img
                                                src={car.image_url || `https://placehold.co/70x50/eee/ccc?text=N/A`}
                                                alt={`${car.brand} ${car.model}`}
                                                style={imgStyle}
                                                onError={(e) => { e.target.onerror = null; e.target.src='https://placehold.co/70x50/eee/ccc?text=Err'; }}
                                            />
                                        </td>
                                        <td>{car.brand}</td>
                                        <td>{car.model}</td>
                                        {/* Format price */}
                                        <td>{car.price_per_day?.toFixed(2)}</td>
                                        <td>{car.branch_id}</td>
                                        <td>
                                            {/* Availability Status Indicator */}
                                            <span className={car.availability ? 'status-available-yes' : 'status-available-no'}>
                                                {car.availability ? 'Yes' : 'No'}
                                            </span>
                                        </td>
                                        <td>{car.parking_spot || '-'}</td>
                                        {/* Action Buttons */}
                                        <td className="actions admin-action-buttons">
                                            <button onClick={() => handleEditCar(car)} className="admin-button admin-button-warning admin-button-sm" disabled={isDeleting}>Edit</button>
                                            <button onClick={() => handleDeleteCar(car.id)} className="admin-button admin-button-danger admin-button-sm" disabled={isDeleting}>
                                                {isDeleting ? '...' : 'Delete'}
                                            </button>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Render CarForm Modal when showForm is true */}
            {showForm && (
                <CarForm initialData={editingCar} onClose={handleCloseForm} />
            )}
        </div>
    );
};

export default CarManagement;
