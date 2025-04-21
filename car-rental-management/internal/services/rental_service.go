package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time" // Needed for time.Time type usage from models

	"github.com/jmoiron/sqlx"
)

// CreateRental creates a new rental record and updates car availability within a transaction
// Accepts the base Rental model, required fields are set before calling.
func CreateRental(rental models.Rental) (models.Rental, error) {
	log.Println("🔍 Validating rental data before creating:", rental)

	// --- Input Validation ---
	if rental.CustomerID <= 0 {
		return models.Rental{}, errors.New("invalid customer ID")
	}
	if rental.CarID <= 0 {
		return models.Rental{}, errors.New("invalid car ID")
	}
	// Check for zero time might be redundant if coming from validated input struct
	if rental.PickupDatetime.IsZero() || rental.DropoffDatetime.IsZero() || !rental.PickupDatetime.Before(rental.DropoffDatetime) {
		return models.Rental{}, errors.New("invalid pickup/dropoff datetime (must be valid and pickup before dropoff)")
	}
	if rental.Status != "Booked" {
		// This should be enforced by the handler logic, but good to double check
		log.Printf("⚠️ Warning: CreateRental called with initial status '%s', forcing to 'Booked'", rental.Status)
		rental.Status = "Booked"
		// Alternatively, return an error:
		// return models.Rental{}, errors.New("internal error: initial rental status must be 'Booked'")
	}
	// --- End Input Validation ---

	log.Println("✅ Rental data validation passed. Attempting to create rental within a transaction...")

	var err error   // Declare error variable accessible within defer
	var tx *sqlx.Tx // Declare transaction variable

	tx, err = config.DB.Beginx()
	if err != nil {
		log.Println("❌ Error starting transaction:", err)
		return models.Rental{}, fmt.Errorf("failed to start database transaction: %w", err)
	}

	// Defer rollback/commit logic
	defer func() {
		if p := recover(); p != nil {
			log.Println("🔥 Panic occurred during rental creation, rolling back transaction:", p)
			_ = tx.Rollback()
			panic(p) // Re-panic after rollback attempt
		} else if err != nil {
			// If an error occurred in the main logic block
			log.Printf("❌ Rolling back rental creation transaction due to error: %v", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("❌ Error during transaction rollback after error: %v", rbErr)
			}
		} else {
			// If no error occurred before commit, attempt commit
			log.Println("⏳ Committing rental creation transaction...")
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Println("❌ Error committing rental creation transaction:", commitErr)
				// Set the outer error variable so the caller knows the commit failed
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			} else {
				log.Println("✅ Rental creation transaction committed successfully!")
			}
		}
	}() // Immediately invoke the deferred function

	// 2. Check car availability within the transaction
	var carAvailable bool
	var carBranchID int
	checkCarQuery := "SELECT availability, branch_id FROM cars WHERE id=$1 FOR UPDATE" // Lock the row and get branch_id
	err = tx.QueryRowx(checkCarQuery, rental.CarID).Scan(&carAvailable, &carBranchID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ Car with ID %d not found for rental.", rental.CarID)
			err = errors.New("car not found") // Set specific error
		} else {
			log.Println("❌ Error checking car availability:", err)
			err = fmt.Errorf("failed to check car availability: %w", err) // Wrap DB error
		}
		return models.Rental{}, err // Trigger rollback by returning error
	}

	if !carAvailable {
		log.Printf("❌ Car with ID %d is not available for rent.", rental.CarID)
		err = errors.New("car is not available") // Set specific error
		return models.Rental{}, err              // Trigger rollback
	}

	// Optional: Add pickup location from branch if not provided by user input
	if rental.PickupLocation == nil || *rental.PickupLocation == "" {
		var branchAddress sql.NullString // Use sql.NullString for potentially null address
		branchQuery := "SELECT address FROM branches WHERE id=$1"
		branchErr := tx.Get(&branchAddress, branchQuery, carBranchID)
		if branchErr != nil {
			// Log warning but don't necessarily fail the rental creation
			log.Printf("⚠️ Could not fetch branch address for car %d (branch %d) to set default pickup: %v", rental.CarID, carBranchID, branchErr)
		} else if branchAddress.Valid && branchAddress.String != "" {
			log.Printf("ℹ️ Setting pickup location automatically from branch %d address.", carBranchID)
			// Assign the valid address string to the rental's PickupLocation pointer
			rental.PickupLocation = &branchAddress.String
		}
	}

	// 3. Insert rental record within the transaction
	insertRentalQuery := `
		INSERT INTO rentals (customer_id, car_id, booking_date, pickup_datetime, dropoff_datetime, pickup_location, status)
		VALUES (:customer_id, :car_id, NOW(), :pickup_datetime, :dropoff_datetime, :pickup_location, :status)
        RETURNING id, created_at, updated_at, booking_date` // Return generated values

	var createdRentalData struct { // Temporary struct to scan returned values
		ID          int
		CreatedAt   time.Time // Need time.Time type here
		UpdatedAt   time.Time
		BookingDate *time.Time // Match model type *time.Time
	}
	stmt, err := tx.PrepareNamed(insertRentalQuery)
	if err != nil {
		log.Println("❌ Error preparing rental insert query:", err)
		err = fmt.Errorf("failed to prepare rental record insert: %w", err)
		return models.Rental{}, err // Trigger rollback
	}
	defer stmt.Close() // Close prepared statement when function exits

	// Execute and scan the returned values using the input `rental` struct for named parameters
	err = stmt.QueryRowx(&rental).Scan(&createdRentalData.ID, &createdRentalData.CreatedAt, &createdRentalData.UpdatedAt, &createdRentalData.BookingDate)
	if err != nil {
		log.Println("❌ Error inserting rental:", err)
		err = fmt.Errorf("failed to create rental record: %w", err)
		return models.Rental{}, err // Trigger rollback
	}
	log.Printf("✅ Rental record %d inserted successfully (within transaction).", createdRentalData.ID)

	// 4. Update car availability to false within the transaction
	updateCarQuery := "UPDATE cars SET availability=false WHERE id=$1"
	_, err = tx.Exec(updateCarQuery, rental.CarID)
	if err != nil {
		log.Println("❌ Error updating car availability:", err)
		err = fmt.Errorf("failed to update car status: %w", err)
		return models.Rental{}, err // Trigger rollback
	}
	log.Println("✅ Car availability updated successfully (within transaction).")

	// Construct the final return object after potential commit
	finalRental := rental
	finalRental.ID = createdRentalData.ID
	finalRental.CreatedAt = createdRentalData.CreatedAt
	finalRental.UpdatedAt = createdRentalData.UpdatedAt
	finalRental.BookingDate = createdRentalData.BookingDate

	if err == nil {
		log.Printf("✅ Rental %d created and car status updated successfully!", finalRental.ID)
	}
	return finalRental, err // Return final error state (nil if commit succeeded)
}

// GetRentals retrieves all rentals (consider filtering/pagination)
func GetRentals() ([]models.Rental, error) {
	var rentals []models.Rental
	log.Println("🔍 Fetching all rentals...")
	query := `SELECT id, customer_id, car_id, booking_date, pickup_datetime, dropoff_datetime, pickup_location, status, created_at, updated_at
              FROM rentals ORDER BY created_at DESC`
	err := config.DB.Select(&rentals, query)
	if err != nil {
		log.Println("❌ Error fetching rentals:", err)
		return nil, fmt.Errorf("failed to fetch rentals: %w", err)
	}
	log.Printf("✅ Fetched %d rentals successfully", len(rentals))
	return rentals, nil
}

// GetRentalByID retrieves a single rental
func GetRentalByID(id int) (models.Rental, error) {
	var rental models.Rental
	log.Println("🔍 Fetching rental by ID:", id)
	if id <= 0 {
		return models.Rental{}, errors.New("invalid rental ID")
	}
	query := `SELECT id, customer_id, car_id, booking_date, pickup_datetime, dropoff_datetime, pickup_location, status, created_at, updated_at
              FROM rentals WHERE id=$1`
	err := config.DB.Get(&rental, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Rental{}, errors.New("rental not found") // Specific error for not found
		}
		log.Printf("❌ Error fetching rental %d: %v", id, err)
		return models.Rental{}, fmt.Errorf("failed to fetch rental %d: %w", id, err) // Wrap other DB errors
	}
	log.Printf("✅ Rental %d fetched successfully", id)
	return rental, nil
}

// UpdateRentalStatus handles status transitions and related actions (like car availability)
func UpdateRentalStatus(rentalID int, newStatus string, employeeID *int) (models.Rental, error) {
	log.Printf("🔄 Attempting to update rental %d status to '%s' by employee %v", rentalID, newStatus, employeeID)

	if rentalID <= 0 {
		return models.Rental{}, errors.New("invalid rental ID")
	}
	allowedStatuses := map[string]bool{"Confirmed": true, "Active": true, "Returned": true, "Cancelled": true}
	if !allowedStatuses[newStatus] {
		return models.Rental{}, fmt.Errorf("invalid target status: %s", newStatus)
	}

	var err error // Error variable for transaction scope
	var tx *sqlx.Tx

	tx, err = config.DB.Beginx()
	if err != nil {
		log.Println("❌ Error starting transaction for status update:", err)
		return models.Rental{}, fmt.Errorf("db transaction error: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			log.Println("🔥 Panic during status update, rolling back:", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Printf("❌ Rolling back status update transaction due to error: %v", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("❌ Error during transaction rollback after error: %v", rbErr)
			}
		} else {
			log.Println("⏳ Committing status update transaction...")
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Println("❌ Error committing status update transaction:", commitErr)
				err = fmt.Errorf("failed to commit status update: %w", commitErr) // Set outer error
			} else {
				log.Println("✅ Status update transaction committed successfully!")
			}
		}
	}() // End defer func

	// 1. Get current rental status and car_id (Lock the row)
	var currentStatus string
	var carID int
	var customerID int
	query := "SELECT status, car_id, customer_id FROM rentals WHERE id=$1 FOR UPDATE" // Lock row
	err = tx.QueryRowx(query, rentalID).Scan(&currentStatus, &carID, &customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("rental not found for status update") // Specific error
		} else {
			log.Println("❌ Error getting current rental status:", err)
			err = fmt.Errorf("failed to get current rental status: %w", err) // Wrap error
		}
		return models.Rental{}, err // Trigger rollback
	}

	// 2. Check if transition is valid
	validTransition := false
	switch currentStatus {
	case "Booked":
		validTransition = (newStatus == "Confirmed" || newStatus == "Cancelled")
	case "Confirmed":
		validTransition = (newStatus == "Active" || newStatus == "Cancelled")
	case "Active":
		validTransition = (newStatus == "Returned")
		if newStatus == "Cancelled" && employeeID != nil {
			log.Printf("ℹ️ Staff (ID: %d) is cancelling an ACTIVE rental (ID: %d).", *employeeID, rentalID)
			validTransition = true
		}
	case "Returned", "Cancelled":
		validTransition = false // Cannot change status after final state
	default:
		validTransition = false // Unknown current status
		log.Printf("⚠️ Unknown current rental status '%s' for rental %d", currentStatus, rentalID)
	}

	if !validTransition {
		err = fmt.Errorf("invalid status transition from '%s' to '%s'", currentStatus, newStatus)
		return models.Rental{}, err // Trigger rollback
	}

	// 3. Update rental status
	log.Printf("Updating rental %d status from %s to %s", rentalID, currentStatus, newStatus)
	updateQuery := "UPDATE rentals SET status=$1 WHERE id=$2"
	result, err := tx.Exec(updateQuery, newStatus, rentalID)
	if err != nil {
		log.Println("❌ Error updating rental status:", err)
		err = fmt.Errorf("failed to update rental status in db: %w", err)
		return models.Rental{}, err // Trigger rollback
	}
	rowsAffected, _ := result.RowsAffected() // Ignore error on RowsAffected if Exec succeeded
	if rowsAffected == 0 {
		err = errors.New("rental not found during status update execution (concurrency issue?)")
		return models.Rental{}, err // Trigger rollback
	}

	// 4. Handle side effects: Update car availability if necessary
	updateCarAvailability := false
	if newStatus == "Returned" {
		updateCarAvailability = true
	} else if newStatus == "Cancelled" && (currentStatus == "Booked" || currentStatus == "Confirmed" || currentStatus == "Active") {
		updateCarAvailability = true
	}

	if updateCarAvailability {
		log.Printf("Updating availability for car %d to true due to rental status change to %s", carID, newStatus)
		carUpdateQuery := "UPDATE cars SET availability=true WHERE id=$1"
		_, err = tx.Exec(carUpdateQuery, carID)
		if err != nil {
			log.Printf("❌ Error updating availability for car %d: %v", carID, err)
			err = fmt.Errorf("failed to update car availability: %w", err)
			return models.Rental{}, err // Trigger rollback
		}
		log.Printf("✅ Marked car %d as available.", carID)
	} else {
		log.Printf("ℹ️ No car availability update needed for status change from '%s' to '%s'", currentStatus, newStatus)
	}

	// ** REMOVED unused 'updatedRental' declaration here **

	// Need to return the updated rental state. Fetch it *after* the defer block runs.
	var returnRental models.Rental // Declare the variable to be returned
	if err == nil {                // If transaction committed successfully
		// Fetch the latest state using the non-transactional function
		fetchRental, fetchErr := GetRentalByID(rentalID)
		if fetchErr != nil {
			log.Printf("⚠️ Could not fetch updated rental %d after status change commit: %v", rentalID, fetchErr)
			// The update *likely* succeeded, but we can't return the full object.
			// Return an error indicating this ambiguity.
			return models.Rental{}, fmt.Errorf("status update committed but failed to fetch result: %w", fetchErr)
		}
		returnRental = fetchRental // Assign fetched data
		log.Printf("✅ Rental %d status updated to %s successfully.", rentalID, newStatus)
	}

	// Return the fetched rental (if successful) and the final transaction error state (nil on success)
	return returnRental, err
}

// DeleteRental removes a rental (use with caution, triggers CASCADE on payments/reviews)
func DeleteRental(rentalID int) error {
	log.Println("🗑 Attempting to delete rental with ID:", rentalID)
	if rentalID <= 0 {
		return errors.New("invalid rental ID")
	}

	result, err := config.DB.Exec("DELETE FROM rentals WHERE id=$1", rentalID)
	if err != nil {
		log.Printf("❌ Error deleting rental %d: %v", rentalID, err)
		return fmt.Errorf("failed to delete rental: %w", err)
	}
	rowsAffected, _ := result.RowsAffected() // Ignore error getting rows affected
	if rowsAffected == 0 {
		return errors.New("rental not found for deletion") // Specific error
	}
	log.Println("✅ Rental deleted successfully!")
	return nil
}

// GetRentalsByCustomerID retrieves rentals for a specific customer
func GetRentalsByCustomerID(customerID int) ([]models.Rental, error) {
	var rentals []models.Rental
	log.Println("🔍 Fetching rentals for customer ID:", customerID)
	if customerID <= 0 {
		return nil, errors.New("invalid customer ID")
	}
	query := `SELECT id, customer_id, car_id, booking_date, pickup_datetime, dropoff_datetime, pickup_location, status, created_at, updated_at
               FROM rentals WHERE customer_id=$1 ORDER BY created_at DESC`
	err := config.DB.Select(&rentals, query, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("ℹ️ No rentals found for customer %d", customerID)
			return []models.Rental{}, nil // Return empty slice, not error
		}
		log.Printf("❌ Error fetching rentals for customer %d: %v", customerID, err)
		return nil, fmt.Errorf("failed to fetch rentals for customer %d: %w", customerID, err)
	}
	log.Printf("✅ Fetched %d rentals for customer %d successfully", len(rentals), customerID)
	return rentals, nil
}

// CancelCustomerRental allows a customer to cancel their own booking (if status allows)
func CancelCustomerRental(rentalID int, customerID int) error {
	log.Printf("Customer %d attempting to cancel rental %d", customerID, rentalID)

	if rentalID <= 0 || customerID <= 0 {
		return errors.New("invalid rental or customer ID")
	}

	var err error // Transaction error variable
	var tx *sqlx.Tx

	tx, err = config.DB.Beginx()
	if err != nil {
		log.Println("❌ Tx Begin Error for customer cancellation:", err)
		return fmt.Errorf("database error starting transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			log.Println("🔥 Panic during customer cancellation, rolling back:", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Printf("❌ Rolling back customer cancellation transaction due to error: %v", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("❌ Error during rollback after error: %v", rbErr)
			}
		} else {
			log.Println("⏳ Committing customer cancellation transaction...")
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Println("❌ Error committing customer cancellation transaction:", commitErr)
				err = fmt.Errorf("failed to commit cancellation: %w", commitErr) // Set outer error
			} else {
				log.Println("✅ Customer cancellation transaction committed successfully!")
			}
		}
	}() // End defer func

	// 1. Get rental details & verify ownership & status (Lock the row)
	var currentStatus string
	var currentCustomerID int
	var carID int
	query := "SELECT status, customer_id, car_id FROM rentals WHERE id=$1 FOR UPDATE" // Lock row
	err = tx.QueryRowx(query, rentalID).Scan(&currentStatus, &currentCustomerID, &carID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("rental not found") // Specific error
		} else {
			log.Printf("❌ DB Error getting rental details for cancellation: %v", err)
			err = fmt.Errorf("failed to get rental details: %w", err) // Wrap DB error
		}
		return err // Trigger rollback
	}

	// Verify ownership
	if currentCustomerID != customerID {
		err = errors.New("permission denied: cannot cancel another customer's rental")
		return err // Trigger rollback
	}

	// 2. Check if cancellation is allowed based on status
	// Customer can only cancel 'Booked' or 'Confirmed' rentals.
	if !(currentStatus == "Booked" || currentStatus == "Confirmed") {
		err = fmt.Errorf("cannot cancel rental with status '%s'", currentStatus)
		return err // Trigger rollback
	}

	// 3. Update status to 'Cancelled'
	updateQuery := "UPDATE rentals SET status='Cancelled' WHERE id=$1"
	result, err := tx.Exec(updateQuery, rentalID)
	if err != nil {
		log.Printf("❌ DB Error updating rental status to Cancelled: %v", err)
		err = fmt.Errorf("failed to update rental status: %w", err)
		return err // Trigger rollback
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Should not happen if fetch+lock worked
		err = errors.New("rental disappeared during cancellation (concurrency issue?)")
		return err // Trigger rollback
	}

	// 4. Update car availability (make it available again)
	carUpdateQuery := "UPDATE cars SET availability=true WHERE id=$1"
	_, err = tx.Exec(carUpdateQuery, carID)
	if err != nil {
		log.Printf("❌ DB Error updating car availability during cancellation: %v", err)
		err = fmt.Errorf("failed to update car availability: %w", err)
		return err // Trigger rollback
	}

	if err == nil {
		log.Printf("✅ Rental %d cancelled successfully by customer %d", rentalID, customerID)
	}
	// Return the final error state (nil if commit succeeded)
	return err
}
