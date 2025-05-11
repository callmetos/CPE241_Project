package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models" // Ensure models is imported
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Define shared error variables for the services package (and for handlers to access via services.ErrXxx)
var (
	ErrRentalNotFound  = errors.New("rental not found")
	ErrCarNotFound     = errors.New("car not found")
	ErrCarNotAvailable = errors.New("car is not available for the selected dates")
	ErrInvalidDates    = errors.New("invalid pickup/dropoff dates")
	ErrForbidden       = errors.New("permission denied")
	ErrInvalidState    = errors.New("invalid operation for current rental/payment state")
)

func InitiateRentalBooking(customerID int, input models.InitiateRentalInput) (models.Rental, error) {
	log.Printf("Service: Initiating rental for customer %d, car %d", customerID, input.CarID)
	if customerID <= 0 {
		return models.Rental{}, errors.New("invalid customer ID")
	}
	if input.CarID <= 0 {
		return models.Rental{}, errors.New("invalid car ID")
	}
	if input.PickupDatetime.IsZero() || input.DropoffDatetime.IsZero() || !input.PickupDatetime.Before(input.DropoffDatetime) || input.PickupDatetime.Before(time.Now().Add(-1*time.Hour)) {
		return models.Rental{}, ErrInvalidDates // Use defined error
	}

	tx, errTx := config.DB.Beginx()
	if errTx != nil {
		log.Printf("❌ InitiateRentalBooking: Failed to begin transaction: %v", errTx)
		return models.Rental{}, fmt.Errorf("database transaction error: %w", errTx)
	}
	var finalErr error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if finalErr != nil {
			log.Printf("❌ Rolling back InitiateRentalBooking tx due to error: %v", finalErr)
			_ = tx.Rollback()
		} else {
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Printf("❌ Error committing InitiateRentalBooking tx: %v", commitErr)
				finalErr = fmt.Errorf("commit error: %w", commitErr)
			} else {
				log.Println("✅ InitiateRentalBooking tx committed.")
			}
		}
	}()

	var car models.Car
	errCar := tx.Get(&car, "SELECT id, availability, branch_id FROM cars WHERE id=$1 FOR UPDATE", input.CarID)
	if errCar != nil {
		if errors.Is(errCar, sql.ErrNoRows) {
			finalErr = ErrCarNotFound // Use defined error
		} else {
			log.Printf("❌ InitiateRentalBooking: Error fetching/locking car %d: %v", input.CarID, errCar)
			finalErr = fmt.Errorf("failed to check car details: %w", errCar)
		}
		return models.Rental{}, finalErr
	}

	var overlapCount int
	overlapQuery := `
            SELECT COUNT(*) FROM rentals
            WHERE car_id = $1
              AND status IN ('Pending', 'Booked', 'Confirmed', 'Active', 'Pending Verification')
              AND (pickup_datetime < $3 AND dropoff_datetime > $2)`
	errOverlap := tx.Get(&overlapCount, overlapQuery, input.CarID, input.PickupDatetime, input.DropoffDatetime)
	if errOverlap != nil {
		log.Printf("❌ InitiateRentalBooking: Error checking for overlapping rentals for car %d: %v", input.CarID, errOverlap)
		finalErr = fmt.Errorf("failed to verify car availability: %w", errOverlap)
		return models.Rental{}, finalErr
	}
	if overlapCount > 0 {
		log.Printf("❌ InitiateRentalBooking: Car %d is not available due to %d overlapping booking(s) for selected dates.", input.CarID, overlapCount)
		finalErr = ErrCarNotAvailable // Use defined error
		return models.Rental{}, finalErr
	}

	rental := models.Rental{
		CustomerID:      customerID,
		CarID:           input.CarID,
		PickupDatetime:  input.PickupDatetime,
		DropoffDatetime: input.DropoffDatetime,
		PickupLocation:  input.PickupLocation,
		Status:          "Pending",
		BookingDate:     nil,
	}

	if rental.PickupLocation == nil || *rental.PickupLocation == "" {
		var branchAddress sql.NullString
		branchQuery := "SELECT address FROM branches WHERE id=$1"
		branchErr := tx.Get(&branchAddress, branchQuery, car.BranchID)
		if branchErr != nil && !errors.Is(branchErr, sql.ErrNoRows) {
			log.Printf("⚠️ InitiateRentalBooking: Could not fetch branch address for car %d (branch %d): %v", rental.CarID, car.BranchID, branchErr)
		} else if branchAddress.Valid && branchAddress.String != "" {
			log.Printf("ℹ️ InitiateRentalBooking: Setting pickup location from branch %d address.", car.BranchID)
			rental.PickupLocation = &branchAddress.String
		} else {
			log.Printf("ℹ️ InitiateRentalBooking: Branch address not found or empty for branch %d. PickupLocation remains as provided.", car.BranchID)
		}
	}

	insertQuery := `
		INSERT INTO rentals (customer_id, car_id, pickup_datetime, dropoff_datetime, pickup_location, status, booking_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at`

	finalErr = tx.QueryRowx(
		insertQuery,
		rental.CustomerID, rental.CarID, rental.PickupDatetime, rental.DropoffDatetime, rental.PickupLocation, rental.Status, rental.BookingDate,
	).Scan(&rental.ID, &rental.CreatedAt, &rental.UpdatedAt)

	if finalErr != nil {
		log.Printf("❌ InitiateRentalBooking: Error inserting pending rental: %v", finalErr)
		finalErr = fmt.Errorf("database error creating pending rental: %w", finalErr)
		return models.Rental{}, finalErr
	}

	log.Printf("✅ Service: Pending rental created with ID: %d", rental.ID)
	return rental, finalErr
}

func GetRentalByID(id int) (models.Rental, error) {
	var rental models.Rental
	log.Println("🔍 Service: Fetching rental by ID:", id)
	if id <= 0 {
		return models.Rental{}, errors.New("invalid rental ID")
	}
	query := `
		SELECT
			r.id, r.customer_id, r.car_id, r.booking_date,
			r.pickup_datetime, r.dropoff_datetime, r.pickup_location,
			r.status, r.created_at, r.updated_at,
			c.brand AS "car.brand",
			c.model AS "car.model"
		FROM rentals r
		JOIN cars c ON r.car_id = c.id
		WHERE r.id=$1`
	err := config.DB.Get(&rental, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Rental{}, ErrRentalNotFound // Use defined error
		}
		log.Printf("❌ Service: Error fetching rental %d: %v", id, err)
		return models.Rental{}, fmt.Errorf("failed to fetch rental %d: %w", id, err)
	}
	log.Printf("✅ Service: Rental %d fetched successfully", id)
	return rental, nil
}

func checkRentalOwnership(rentalID, customerID int) (models.Rental, error) {
	rental, err := GetRentalByID(rentalID)
	if err != nil {
		return models.Rental{}, err
	}
	if rental.CustomerID != customerID {
		log.Printf("🚫 Service: Customer %d does not own Rental %d (Owner: %d)", customerID, rentalID, rental.CustomerID)
		return rental, ErrForbidden // Use defined error
	}
	return rental, nil
}

// UpdateRentalStatus updates a rental's status and handles car availability.
// It can operate within an existing transaction (tx) or create its own if tx is nil.
func UpdateRentalStatus(tx *sqlx.Tx, rentalID int, newStatus string, employeeID *int) (updatedRental models.Rental, err error) {
	log.Printf("🔄 Service: Attempting to update rental %d status to '%s' (Employee: %v)", rentalID, newStatus, employeeID)

	if rentalID <= 0 {
		err = errors.New("invalid rental ID")
		return
	}
	allowedStatuses := map[string]bool{
		"Booked": true, "Confirmed": true, "Active": true, "Returned": true, "Cancelled": true, "Failed": true, "Pending Verification": true,
	}
	if !allowedStatuses[newStatus] {
		err = fmt.Errorf("invalid target status for UpdateRentalStatus function: %s", newStatus)
		return
	}

	// Determine if this function is responsible for managing the transaction
	var currentTx *sqlx.Tx
	responsibleForCommit := false
	if tx == nil {
		currentTx, err = config.DB.Beginx()
		if err != nil {
			log.Println("❌ Tx Begin Error in UpdateRentalStatus:", err)
			err = fmt.Errorf("db transaction error: %w", err)
			return
		}
		responsibleForCommit = true // This function instance will commit/rollback
		defer func() {
			if p := recover(); p != nil {
				log.Println("🔥 Panic during status update, rolling back:", p)
				_ = currentTx.Rollback()
				panic(p)
			} else if err != nil {
				log.Printf("❌ Rolling back status update tx due to error (UpdateRentalStatus): %v", err)
				_ = currentTx.Rollback()
			} else if responsibleForCommit { // Only commit if this instance started the tx
				log.Println("⏳ Committing status update transaction (UpdateRentalStatus)...")
				commitErr := currentTx.Commit()
				if commitErr != nil {
					log.Println("❌ Error committing status update tx (UpdateRentalStatus):", commitErr)
					err = fmt.Errorf("commit error: %w", commitErr)
				} else {
					log.Println("✅ Status update tx committed (UpdateRentalStatus).")
				}
			}
		}()
	} else {
		currentTx = tx // Use the transaction passed in
	}

	var currentStatus string
	var carID int
	// Use currentTx for all DB operations within this function
	query := "SELECT status, car_id FROM rentals WHERE id=$1 FOR UPDATE"
	err = currentTx.QueryRowx(query, rentalID).Scan(&currentStatus, &carID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrRentalNotFound
		} else {
			err = fmt.Errorf("db error getting current status: %w", err)
		}
		return
	}

	validTransition := false
	switch currentStatus {
	case "Pending":
		validTransition = (newStatus == "Booked" || newStatus == "Cancelled" || newStatus == "Pending Verification")
	case "Booked":
		validTransition = (newStatus == "Confirmed" || newStatus == "Cancelled" || newStatus == "Failed" || newStatus == "Pending Verification")
	case "Pending Verification":
		validTransition = (newStatus == "Confirmed" || newStatus == "Failed" || newStatus == "Cancelled")
	case "Confirmed":
		validTransition = (newStatus == "Active" || newStatus == "Cancelled")
	case "Active":
		validTransition = (newStatus == "Returned" || newStatus == "Cancelled")
	case "Returned", "Cancelled", "Failed":
		validTransition = false
		log.Printf("ℹ️ Rental %d is already in a terminal state '%s'. Cannot transition to '%s'.", rentalID, currentStatus, newStatus)
	default:
		validTransition = false
		log.Printf("⚠️ Unknown current status '%s' for rental %d", currentStatus, rentalID)
	}

	if !validTransition {
		err = fmt.Errorf("invalid status transition from '%s' to '%s': %w", currentStatus, newStatus, ErrInvalidState)
		return
	}

	log.Printf("Updating rental %d status from %s to %s", rentalID, currentStatus, newStatus)
	updateQuery := "UPDATE rentals SET status=$1, updated_at = NOW()"
	if newStatus == "Confirmed" || newStatus == "Booked" { // Ensure booking_date is set upon confirmation/booking
		updateQuery += ", booking_date = COALESCE(booking_date, NOW())"
	}
	updateQuery += " WHERE id=$2"

	result, execErr := currentTx.Exec(updateQuery, newStatus, rentalID)
	if execErr != nil {
		err = fmt.Errorf("db error updating rental status: %w", execErr)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		err = ErrRentalNotFound
		return
	}

	var newCarAvailability bool
	updateCarFlag := false

	switch newStatus {
	case "Booked", "Confirmed", "Active", "Pending Verification":
		newCarAvailability = false
		updateCarFlag = true
	case "Returned", "Cancelled", "Failed":
		newCarAvailability = true
		updateCarFlag = true
	}

	if updateCarFlag {
		if newCarAvailability { // If trying to set car to available
			var activeBookingCount int
			// Check for other bookings that would keep the car unavailable
			checkOverlapQuery := `SELECT COUNT(*) FROM rentals
                                  WHERE car_id = $1 AND id != $2
                                  AND status IN ('Booked', 'Confirmed', 'Active', 'Pending Verification')
                                  AND dropoff_datetime > NOW()` // Consider only future or ongoing bookings
			countErr := currentTx.Get(&activeBookingCount, checkOverlapQuery, carID, rentalID)
			if countErr != nil && !errors.Is(countErr, sql.ErrNoRows) {
				log.Printf("⚠️ UpdateRentalStatus: Failed to check other active bookings for car %d: %v", carID, countErr)
				// Optionally, proceed without this check if it fails, or make it critical
			}
			if activeBookingCount > 0 {
				log.Printf("ℹ️ UpdateRentalStatus: Car %d still has %d other active/confirmed/booked rentals. Keeping availability as false.", carID, activeBookingCount)
				newCarAvailability = false // Override, car is still needed
			}
		}

		log.Printf("Updating availability for car %d to %t due to rental status change to %s", carID, newCarAvailability, newStatus)
		carUpdateQuery := "UPDATE cars SET availability=$1, updated_at = NOW() WHERE id=$2"
		_, carUpdateErr := currentTx.Exec(carUpdateQuery, newCarAvailability, carID)
		if carUpdateErr != nil {
			// This is a critical part of the transaction. If it fails, the whole thing should ideally roll back.
			log.Printf("❌ UpdateRentalStatus: CRITICAL - Failed to update car availability for car %d to %t: %v. Rental status updated to %s.", carID, newCarAvailability, carUpdateErr, newStatus)
			err = fmt.Errorf("failed to update car availability: %w", carUpdateErr) // Set error to trigger rollback
			return
		} else {
			log.Printf("✅ Marked car %d availability as %t.", carID, newCarAvailability)
		}
	} else {
		log.Printf("ℹ️ No car availability update needed for status change from '%s' to '%s'", currentStatus, newStatus)
	}

	// Fetch the updated rental details (outside the transaction if this function manages its own)
	// If called with an existing tx, this Get should also use that tx for consistency.
	// However, GetRentalByID creates its own connection. For simplicity for this example, we fetch after commit.
	// A better approach for GetRentalByID would also be to accept an optional tx.
	if err == nil { // Only fetch if no preceding errors
		// If this function is responsible for commit, fetch *after* potential commit.
		// If using a passed-in tx, the caller will commit, so fetching here with the same tx is fine.
		var tempRental models.Rental
		// Construct a query to get rental and car summary within the current transaction
		// to avoid issues with GetRentalByID opening a new connection.
		fetchQuery := `
            SELECT
                r.id, r.customer_id, r.car_id, r.booking_date,
                r.pickup_datetime, r.dropoff_datetime, r.pickup_location,
                r.status, r.created_at, r.updated_at,
                c.brand AS "car.brand",
                c.model AS "car.model"
            FROM rentals r
            JOIN cars c ON r.car_id = c.id
            WHERE r.id=$1`
		fetchErr := currentTx.Get(&tempRental, fetchQuery, rentalID)
		if fetchErr != nil {
			log.Printf("⚠️ Could not fetch updated rental %d details after status change (within tx or after commit): %v", rentalID, fetchErr)
			if err == nil { // Don't overwrite a more critical preceding error
				err = fmt.Errorf("status update potentially successful but failed to retrieve updated rental details: %w", fetchErr)
			}
		} else {
			updatedRental = tempRental
		}
	}
	return
}

func GetRentalsPaginated(filters models.RentalFiltersWithPagination) (models.PaginatedRentalsResponse, error) {
	var response models.PaginatedRentalsResponse
	response.Rentals = []models.Rental{}

	queryBuilder := strings.Builder{}
	countQueryBuilder := strings.Builder{}
	args := []interface{}{}
	paramCount := 1

	baseSelect := `
		SELECT
			r.id, r.customer_id, r.car_id, r.booking_date,
			r.pickup_datetime, r.dropoff_datetime, r.pickup_location,
			r.status, r.created_at, r.updated_at,
			c.brand AS "car.brand",
			c.model AS "car.model"
		FROM rentals r
		JOIN cars c ON r.car_id = c.id`
	queryBuilder.WriteString(baseSelect)
	countQueryBuilder.WriteString("SELECT COUNT(r.id) FROM rentals r JOIN cars c ON r.car_id = c.id")

	var conditions []string
	if filters.RentalID != nil && *filters.RentalID > 0 {
		conditions = append(conditions, fmt.Sprintf("r.id = $%d", paramCount))
		args = append(args, *filters.RentalID)
		paramCount++
	}
	if filters.CustomerID != nil && *filters.CustomerID > 0 {
		conditions = append(conditions, fmt.Sprintf("r.customer_id = $%d", paramCount))
		args = append(args, *filters.CustomerID)
		paramCount++
	}
	if filters.CarID != nil && *filters.CarID > 0 {
		conditions = append(conditions, fmt.Sprintf("r.car_id = $%d", paramCount))
		args = append(args, *filters.CarID)
		paramCount++
	}
	if filters.Status != nil && *filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("r.status ILIKE $%d", paramCount))
		args = append(args, *filters.Status)
		paramCount++
	}
	if filters.PickupDateAfter != nil && !filters.PickupDateAfter.IsZero() {
		conditions = append(conditions, fmt.Sprintf("r.pickup_datetime >= $%d", paramCount))
		args = append(args, *filters.PickupDateAfter)
		paramCount++
	}

	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		queryBuilder.WriteString(whereClause)
		countQueryBuilder.WriteString(whereClause)
	}

	err := config.DB.QueryRow(countQueryBuilder.String(), args...).Scan(&response.TotalCount)
	if err != nil {
		log.Printf("Error counting rentals: %v", err)
		return response, fmt.Errorf("failed to count rentals: %w", err)
	}

	orderByClause := " ORDER BY r.id DESC"
	if filters.SortBy != "" {
		validSortByFields := map[string]bool{
			"id": true, "customer_id": true, "car_id": true, "status": true,
			"pickup_datetime": true, "dropoff_datetime": true, "booking_date": true, "created_at": true,
		}
		if validSortByFields[filters.SortBy] {
			sortDir := "ASC"
			if strings.ToUpper(filters.SortDirection) == "DESC" {
				sortDir = "DESC"
			}
			orderByClause = fmt.Sprintf(" ORDER BY r.%s %s, r.id %s", filters.SortBy, sortDir, sortDir)
		}
	}
	queryBuilder.WriteString(orderByClause)

	if filters.Limit <= 0 {
		filters.Limit = 10
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}
	offset := (filters.Page - 1) * filters.Limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCount, paramCount+1))
	args = append(args, filters.Limit, offset)

	rows, err := config.DB.Queryx(queryBuilder.String(), args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, nil
		}
		log.Printf("Error fetching paginated rentals: %v", err)
		return response, fmt.Errorf("failed to fetch rentals: %w", err)
	}
	defer rows.Close()

	var tempRentals []models.Rental
	for rows.Next() {
		var rental models.Rental
		if errScan := rows.StructScan(&rental); errScan != nil {
			log.Printf("Error scanning rental: %v", errScan)
			continue
		}
		tempRentals = append(tempRentals, rental)
	}
	if errRows := rows.Err(); errRows != nil {
		log.Printf("Error iterating rental rows: %v", errRows)
		return response, fmt.Errorf("error processing rental rows: %w", errRows)
	}
	response.Rentals = tempRentals

	response.Page = filters.Page
	response.Limit = filters.Limit
	if response.TotalCount > 0 && response.Limit > 0 {
		response.TotalPages = int(math.Ceil(float64(response.TotalCount) / float64(response.Limit)))
	} else {
		response.TotalPages = 0
	}
	log.Printf("Service: Fetched %d rentals. Page: %d, Limit: %d, TotalItems: %d, TotalPages: %d", len(response.Rentals), response.Page, response.Limit, response.TotalCount, response.TotalPages)
	return response, nil
}

func GetRentals() ([]models.Rental, error) {
	paginatedFilters := models.RentalFiltersWithPagination{
		Page:          1,
		Limit:         10000,
		SortBy:        "id",
		SortDirection: "DESC",
	}
	result, err := GetRentalsPaginated(paginatedFilters)
	if err != nil {
		return nil, err
	}
	return result.Rentals, nil
}

func DeleteRental(rentalID int) error {
	log.Println("🗑 Service: Deleting rental with ID:", rentalID)
	if rentalID <= 0 {
		return errors.New("invalid rental ID")
	}
	result, err := config.DB.Exec("DELETE FROM rentals WHERE id=$1", rentalID)
	if err != nil {
		log.Printf("❌ Service: Error deleting rental %d: %v", rentalID, err)
		return fmt.Errorf("failed to delete rental: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRentalNotFound // Use defined error
	}
	log.Println("✅ Service: Rental deleted successfully!")
	return nil
}

func GetRentalsByCustomerIDPaginated(customerID int, page int, limit int) (models.PaginatedRentalsResponse, error) {
	if customerID <= 0 {
		return models.PaginatedRentalsResponse{}, errors.New("invalid customer ID")
	}
	filters := models.RentalFiltersWithPagination{
		CustomerID:    &customerID,
		Page:          page,
		Limit:         limit,
		SortBy:        "pickup_datetime",
		SortDirection: "DESC",
	}
	return GetRentalsPaginated(filters)
}

func GetRentalsByCustomerID(customerID int) ([]models.Rental, error) {
	response, err := GetRentalsByCustomerIDPaginated(customerID, 1, 100) // Default to fetching a large number
	if err != nil {
		return nil, err
	}
	return response.Rentals, nil
}

func CancelCustomerRental(rentalID int, customerID int) error {
	log.Printf("Service: Customer %d attempting to cancel rental %d", customerID, rentalID)
	if rentalID <= 0 || customerID <= 0 {
		return errors.New("invalid rental or customer ID")
	}

	rental, err := checkRentalOwnership(rentalID, customerID)
	if err != nil {
		return err // ErrForbidden or ErrRentalNotFound from checkRentalOwnership
	}

	// Allow cancellation if Pending, Booked, Confirmed, or Pending Verification
	if !(rental.Status == "Pending" || rental.Status == "Booked" || rental.Status == "Confirmed" || rental.Status == "Pending Verification") {
		return fmt.Errorf("cannot cancel rental with status '%s': %w", rental.Status, ErrInvalidState) // Use defined error
	}

	// UpdateRentalStatus will handle its own transaction or use a passed one.
	// Since we are not in a transaction here, it will create its own.
	_, err = UpdateRentalStatus(nil, rentalID, "Cancelled", nil) // Pass nil for tx
	if err != nil {
		return fmt.Errorf("failed to process cancellation: %w", err)
	}

	log.Printf("✅ Service: Rental %d cancelled successfully by customer %d", rentalID, customerID)
	return nil
}

func CalculateRentalCost(rentalID int) (models.Payment, error) {
	log.Println("Calculating cost for rental ID:", rentalID)
	if rentalID <= 0 {
		return models.Payment{}, errors.New("invalid rental ID")
	}

	var rentalData struct {
		Pickup  time.Time `db:"pickup_datetime"`
		Dropoff time.Time `db:"dropoff_datetime"`
		Status  string    `db:"status"`
		Price   float64   `db:"price_per_day"`
		CarID   int       `db:"car_id"`
	}

	query := `SELECT r.pickup_datetime, r.dropoff_datetime, r.status, c.price_per_day, r.car_id
              FROM rentals r
              JOIN cars c ON r.car_id = c.id
              WHERE r.id=$1`
	err := config.DB.Get(&rentalData, query, rentalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Payment{}, ErrRentalNotFound
		}
		log.Printf("❌ CalculateRentalCost: DB error getting rental/car data for rental %d: %v", rentalID, err)
		return models.Payment{}, fmt.Errorf("db error getting rental/car data: %w", err)
	}

	if !rentalData.Dropoff.After(rentalData.Pickup) {
		log.Printf("❌ CalculateRentalCost: Invalid dates for rental %d: Pickup %v, Dropoff %v", rentalID, rentalData.Pickup, rentalData.Dropoff)
		return models.Payment{}, ErrInvalidDates
	}
	if rentalData.Price <= 0 {
		log.Printf("❌ CalculateRentalCost: Invalid car price (%.2f) for car %d", rentalData.Price, rentalData.CarID)
		return models.Payment{}, errors.New("invalid car price")
	}

	duration := rentalData.Dropoff.Sub(rentalData.Pickup)
	hours := duration.Hours()

	if hours <= 0 { // Less than or equal to 0 hours means no rental period
		log.Printf("⚠️ CalculateRentalCost: Duration is zero or negative for rental %d. Setting cost to 0.", rentalID)
		// Create a payment object with 0 amount. Status might depend on business rules.
		return models.Payment{RentalID: rentalID, Amount: 0, PaymentStatus: "Pending"}, nil // Or another appropriate status
	}

	// Calculate days: Ceil to the next full day.
	// Example: 1 hour = 1 day, 25 hours = 2 days
	rentalDays := int(math.Ceil(hours / 24.0))
	if rentalDays == 0 && hours > 0 { // If duration is > 0 but < 24h, count as 1 day
		rentalDays = 1
	}

	baseCost := float64(rentalDays) * rentalData.Price
	vatRate := 0.07 // 7% VAT
	vatAmount := baseCost * vatRate
	totalCost := baseCost + vatAmount

	log.Printf("✅ Calculated cost for rental %d (%d days, %.2f hours): Total %.2f (Base: %.2f, VAT: %.2f)",
		rentalID, rentalDays, hours, totalCost, baseCost, vatAmount)

	return models.Payment{
		RentalID: rentalID,
		Amount:   totalCost, // This is the total amount including VAT
		// Fields like base_price, vat_amount can be added to models.Payment if needed for breakdown
		PaymentStatus: "Pending", // Default status for calculation
	}, nil
}
