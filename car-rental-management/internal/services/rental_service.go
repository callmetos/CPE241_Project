package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx" // Needed for transaction methods like QueryRowx
)

// CreateRental creates a new rental record and updates car availability within a transaction
func CreateRental(rental models.Rental) (models.Rental, error) { // Return created rental
	log.Println("🔍 Validating rental data before creating:", rental)

	// --- Input Validation ---
	if rental.CustomerID <= 0 {
		return models.Rental{}, errors.New("invalid customer ID")
	}
	if rental.CarID <= 0 {
		return models.Rental{}, errors.New("invalid car ID")
	}
	if rental.PickupDatetime.IsZero() || rental.DropoffDatetime.IsZero() || !rental.PickupDatetime.Before(rental.DropoffDatetime) {
		return models.Rental{}, errors.New("invalid pickup/dropoff datetime")
	}
	// Ensure initial status is valid for creation
	if rental.Status != "Booked" { // Enforce starting status?
		return models.Rental{}, errors.New("initial rental status must be 'Booked'")
	}
	// --- End Input Validation ---

	log.Println("✅ Rental data validation passed. Attempting to create rental within a transaction...")

	var err error   // Declare error variable for defer scope
	var tx *sqlx.Tx // Declare transaction variable

	// 1. Start a transaction
	tx, err = config.DB.Beginx()
	if err != nil {
		log.Println("❌ Error starting transaction:", err)
		return models.Rental{}, errors.New("failed to start database transaction")
	}

	// Defer rollback/commit logic
	defer func() {
		if p := recover(); p != nil {
			log.Println("🔥 Panic occurred, rolling back transaction:", p)
			_ = tx.Rollback()
			panic(p) // Re-panic after rollback
		} else if err != nil {
			log.Println("❌ Rolling back transaction due to error:", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Println("❌ Error during transaction rollback:", rbErr)
			}
		} else {
			// No error before commit, attempt commit
			log.Println("⏳ Committing transaction...")
			err = tx.Commit() // Assign commit error back to err
			if err != nil {
				log.Println("❌ Error committing transaction:", err)
				// Important: Set error so function returns it
				err = errors.New("failed to commit transaction")
			} else {
				log.Println("✅ Transaction committed successfully!")
			}
		}
	}() // Immediately invoke the deferred function

	// 2. Check car availability within the transaction
	var carAvailable bool
	checkCarQuery := "SELECT availability FROM cars WHERE id=$1 FOR UPDATE" // Lock the row
	err = tx.Get(&carAvailable, checkCarQuery, rental.CarID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ Car with ID %d not found for rental.", rental.CarID)
			err = errors.New("car not found")
		} else {
			log.Println("❌ Error checking car availability:", err)
			err = errors.New("failed to check car availability")
		}
		return models.Rental{}, err // Trigger rollback
	}

	if !carAvailable {
		log.Printf("❌ Car with ID %d is not available for rent.", rental.CarID)
		err = errors.New("car is not available")
		return models.Rental{}, err // Trigger rollback
	}

	// 3. Insert rental record within the transaction
	// Use NamedExec for inserting struct directly
	insertRentalQuery := `
		INSERT INTO rentals (customer_id, car_id, booking_date, pickup_datetime, dropoff_datetime, pickup_location, status)
		VALUES (:customer_id, :car_id, :booking_date, :pickup_datetime, :dropoff_datetime, :pickup_location, :status)
        RETURNING id, created_at, updated_at, booking_date` // Return generated values
	// We need to use QueryRowx or similar to scan back the generated values with NamedExec parameters
	// Let's prepare the statement
	var createdRental models.Rental // To store the final rental data
	stmt, err := tx.PrepareNamed(insertRentalQuery)
	if err != nil {
		log.Println("❌ Error preparing rental insert query:", err)
		err = errors.New("failed to prepare rental record insert")
		return models.Rental{}, err // Trigger rollback
	}
	defer stmt.Close()
	// Execute and scan the returned values
	err = stmt.QueryRowx(rental).Scan(&createdRental.ID, &createdRental.CreatedAt, &createdRental.UpdatedAt, &createdRental.BookingDate)

	if err != nil {
		log.Println("❌ Error inserting rental:", err)
		err = errors.New("failed to create rental record")
		return models.Rental{}, err // Trigger rollback
	}
	log.Printf("✅ Rental record %d inserted successfully (within transaction).", createdRental.ID)

	// 4. Update car availability to false within the transaction
	updateCarQuery := "UPDATE cars SET availability=false WHERE id=$1"
	_, err = tx.Exec(updateCarQuery, rental.CarID)
	if err != nil {
		log.Println("❌ Error updating car availability:", err)
		err = errors.New("failed to update car status")
		return models.Rental{}, err // Trigger rollback
	}
	log.Println("✅ Car availability updated successfully (within transaction).")

	// If we reach here without error, the deferred function will commit.
	log.Printf("✅ Rental %d created and car status updated successfully!", createdRental.ID)

	// Populate the rest of the returned struct from input
	createdRental.CustomerID = rental.CustomerID
	createdRental.CarID = rental.CarID
	createdRental.PickupDatetime = rental.PickupDatetime
	createdRental.DropoffDatetime = rental.DropoffDatetime
	createdRental.PickupLocation = rental.PickupLocation
	createdRental.Status = rental.Status // Should be "Booked"

	return createdRental, err // err will be nil if commit succeeded, or commit error otherwise
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
		return nil, errors.New("failed to fetch rentals")
	}
	log.Printf("✅ Fetched %d rentals successfully", len(rentals))
	return rentals, nil
}

// GetRentalByID retrieves a single rental
func GetRentalByID(id int) (models.Rental, error) {
	var rental models.Rental
	log.Println("🔍 Fetching rental by ID:", id)
	query := `SELECT id, customer_id, car_id, booking_date, pickup_datetime, dropoff_datetime, pickup_location, status, created_at, updated_at
              FROM rentals WHERE id=$1`
	err := config.DB.Get(&rental, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Rental{}, errors.New("rental not found")
		}
		log.Printf("❌ Error fetching rental %d: %v", id, err)
		return models.Rental{}, errors.New("failed to fetch rental")
	}
	log.Printf("✅ Rental %d fetched successfully", id)
	return rental, nil
}

// UpdateRentalStatus handles status transitions and related actions (like car availability)
func UpdateRentalStatus(rentalID int, newStatus string, employeeID *int) (models.Rental, error) { // employeeID is pointer for optional logging/checks
	log.Printf("🔄 Attempting to update rental %d status to '%s' by employee %v", rentalID, newStatus, employeeID)

	// Validate newStatus against allowed values (already done by DB CHECK, but good practice)
	allowedStatuses := map[string]bool{"Confirmed": true, "Active": true, "Returned": true, "Cancelled": true}
	if !allowedStatuses[newStatus] {
		return models.Rental{}, fmt.Errorf("invalid target status: %s", newStatus)
	}

	var err error
	var tx *sqlx.Tx

	tx, err = config.DB.Beginx()
	if err != nil {
		log.Println("❌ Error starting transaction for status update:", err)
		return models.Rental{}, errors.New("db transaction error")
	}
	// Defer rollback/commit logic
	defer func() {
		if p := recover(); p != nil {
			log.Println("🔥 Panic during status update, rolling back:", p)
			_ = tx.Rollback()
			panic(p) // Re-panic after rollback
		} else if err != nil {
			log.Println("❌ Rolling back status update due to error:", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Println("❌ Error during transaction rollback:", rbErr)
			}
		} else {
			// No error before commit, attempt commit
			log.Println("⏳ Committing status update transaction...")
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Println("❌ Error committing status update transaction:", commitErr)
				err = errors.New("failed to commit status update") // Set outer error to be returned
			} else {
				log.Println("✅ Status update transaction committed successfully!")
			}
		}
	}() // End defer func

	// 1. Get current rental status and car_id (Lock the row)
	var currentStatus string
	var carID int
	var customerID int // Get customer ID too, might be useful
	query := "SELECT status, car_id, customer_id FROM rentals WHERE id=$1 FOR UPDATE"
	err = tx.QueryRowx(query, rentalID).Scan(&currentStatus, &carID, &customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("rental not found for status update")
		} else {
			log.Println("❌ Error getting current rental status:", err)
			err = errors.New("failed to get current rental status")
		}
		return models.Rental{}, err // Trigger rollback
	}

	// 2. Check if transition is valid (More robust logic)
	validTransition := false
	switch currentStatus {
	case "Booked":
		validTransition = (newStatus == "Confirmed" || newStatus == "Cancelled")
	case "Confirmed":
		validTransition = (newStatus == "Active" || newStatus == "Cancelled")
	case "Active":
		// Typically only moves to Returned. Cancellation of Active might need special handling/permissions.
		validTransition = (newStatus == "Returned")
		// Example: Allow staff (employeeID is not nil) to force cancel an Active rental
		if newStatus == "Cancelled" && employeeID != nil {
			log.Printf("ℹ️ Staff (ID: %d) is cancelling an ACTIVE rental (ID: %d).", *employeeID, rentalID)
			validTransition = true
		}
	case "Returned", "Cancelled":
		validTransition = false // Cannot change status after final state
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
		err = errors.New("failed to update rental status in db")
		return models.Rental{}, err // Trigger rollback
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Should not happen if fetch worked, but good to check
		err = errors.New("rental not found during status update execution")
		return models.Rental{}, err // Trigger rollback
	}

	// 4. Handle side effects: Update car availability if necessary
	updateCarAvailability := false
	// Make car available if rental is Returned OR if it's Cancelled from a state where the car was reserved/out
	if newStatus == "Returned" {
		updateCarAvailability = true
	} else if newStatus == "Cancelled" && (currentStatus == "Booked" || currentStatus == "Confirmed" || currentStatus == "Active") {
		// If cancelling an active rental, car becomes available.
		// If cancelling Booked/Confirmed, car becomes available.
		updateCarAvailability = true
	}

	if updateCarAvailability {
		log.Printf("Updating availability for car %d to true due to rental status change to %s", carID, newStatus)
		carUpdateQuery := "UPDATE cars SET availability=true WHERE id=$1"
		_, err = tx.Exec(carUpdateQuery, carID)
		if err != nil {
			log.Printf("❌ Error updating availability for car %d: %v", carID, err)
			err = errors.New("failed to update car availability")
			return models.Rental{}, err // Trigger rollback
		}
		log.Printf("✅ Marked car %d as available.", carID)
	} else {
		log.Printf("ℹ️ No car availability update needed for status change from '%s' to '%s'", currentStatus, newStatus)
	}

	// If all successful, err remains nil, defer will commit
	if err == nil {
		log.Printf("✅ Rental %d status updated to %s successfully.", rentalID, newStatus)
	}

	// Fetch the updated rental data AFTER potential commit
	// Need to handle the case where err is non-nil due to commit failure
	var updatedRental models.Rental
	if err == nil { // Only attempt fetch if commit likely succeeded
		updatedRental, err = GetRentalByID(rentalID) // Use existing function
		if err != nil {
			log.Printf("⚠️ Could not fetch updated rental %d after status change: %v", rentalID, err)
			// Return minimal info indicating status change likely happened
			// Or return the commit error if that's what 'err' holds now
			return models.Rental{ID: rentalID, Status: newStatus}, err
		}
	}

	return updatedRental, err // Returns the updated rental on success, or the error encountered
}

// DeleteRental removes a rental (use with caution)
func DeleteRental(rentalID int) error {
	log.Println("🗑 Deleting rental with ID:", rentalID)
	// Consider implications: Payments/Reviews are set to CASCADE delete.
	// Car/Customer are RESTRICT delete by default in schema.
	// Should check status before allowing delete? e.g., only delete 'Cancelled'?
	result, err := config.DB.Exec("DELETE FROM rentals WHERE id=$1", rentalID)
	if err != nil {
		log.Printf("❌ Error deleting rental %d: %v", rentalID, err)
		// Check for FK violation if RESTRICT was used differently
		return errors.New("failed to delete rental")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("rental not found for deletion")
	}
	log.Println("✅ Rental deleted successfully!")
	return nil
}

// GetRentalsByCustomerID retrieves rentals for a specific customer
func GetRentalsByCustomerID(customerID int) ([]models.Rental, error) {
	var rentals []models.Rental
	log.Println("🔍 Fetching rentals for customer ID:", customerID)
	query := `SELECT id, customer_id, car_id, booking_date, pickup_datetime, dropoff_datetime, pickup_location, status, created_at, updated_at
               FROM rentals WHERE customer_id=$1 ORDER BY created_at DESC`
	err := config.DB.Select(&rentals, query, customerID)
	if err != nil {
		log.Printf("❌ Error fetching rentals for customer %d: %v", customerID, err)
		return nil, errors.New("failed to fetch rentals")
	}
	log.Printf("✅ Fetched %d rentals for customer %d successfully", len(rentals), customerID)
	return rentals, nil
}

// CancelCustomerRental allows a customer to cancel their own booking (if allowed)
func CancelCustomerRental(rentalID int, customerID int) error {
	log.Printf("Customer %d attempting to cancel rental %d", customerID, rentalID)

	var err error
	var tx *sqlx.Tx

	tx, err = config.DB.Beginx()
	if err != nil {
		log.Println("❌ Tx Begin Error:", err)
		return errors.New("database error")
	}
	defer func() { /* handle rollback/commit */ }() // Implement defer

	// 1. Get rental details & verify ownership & status
	var currentStatus string
	var currentCustomerID int
	var carID int
	query := "SELECT status, customer_id, car_id FROM rentals WHERE id=$1 FOR UPDATE"
	err = tx.QueryRowx(query, rentalID).Scan(&currentStatus, &currentCustomerID, &carID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("rental not found")
			return err
		}
		err = errors.New("failed to get rental details")
		return err
	}

	if currentCustomerID != customerID {
		err = errors.New("permission denied: cannot cancel another customer's rental")
		return err
	}

	// 2. Check if cancellation is allowed based on status
	if !(currentStatus == "Booked" || currentStatus == "Confirmed") {
		err = fmt.Errorf("cannot cancel rental with status '%s'", currentStatus)
		return err
	}

	// 3. Update status to 'Cancelled'
	updateQuery := "UPDATE rentals SET status='Cancelled' WHERE id=$1"
	_, err = tx.Exec(updateQuery, rentalID)
	if err != nil {
		err = errors.New("failed to update rental status")
		return err
	}

	// 4. Update car availability
	carUpdateQuery := "UPDATE cars SET availability=true WHERE id=$1"
	_, err = tx.Exec(carUpdateQuery, carID)
	if err != nil {
		err = errors.New("failed to update car availability")
		return err
	}

	// If successful, defer will commit
	log.Printf("✅ Rental %d cancelled successfully by customer %d", rentalID, customerID)
	return err // Return commit error if any
}
