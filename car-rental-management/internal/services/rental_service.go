package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

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

	if input.PickupDatetime.IsZero() || input.DropoffDatetime.IsZero() || !input.PickupDatetime.Before(input.DropoffDatetime) || input.PickupDatetime.Before(time.Now()) {
		return models.Rental{}, ErrInvalidDates
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
			finalErr = ErrCarNotFound
		} else {
			log.Printf("❌ InitiateRentalBooking: Error fetching/locking car %d: %v", input.CarID, errCar)
			finalErr = fmt.Errorf("failed to check car details: %w", errCar)
		}
		return models.Rental{}, finalErr
	}
	if !car.Availability {
		var overlapCount int
		overlapQuery := `
            SELECT COUNT(*) FROM rentals
            WHERE car_id = $1
              AND status IN ('Booked', 'Confirmed', 'Active')
              AND (pickup_datetime < $3 AND dropoff_datetime > $2)`
		errOverlap := tx.Get(&overlapCount, overlapQuery, input.CarID, input.PickupDatetime, input.DropoffDatetime)
		if errOverlap != nil {
			log.Printf("❌ InitiateRentalBooking: Error checking for overlapping rentals for car %d: %v", input.CarID, errOverlap)
			finalErr = fmt.Errorf("failed to verify car availability: %w", errOverlap)
			return models.Rental{}, finalErr
		}
		if overlapCount > 0 {
			log.Printf("❌ InitiateRentalBooking: Car %d is not available due to %d overlapping booking(s) for selected dates.", input.CarID, overlapCount)
			finalErr = ErrCarNotAvailable
			return models.Rental{}, finalErr
		}
		log.Printf("⚠️ InitiateRentalBooking: Car %d has Availability=false but no overlaps found. Proceeding cautiously.", input.CarID)
	} else {
		var overlapCount int
		overlapQuery := `
            SELECT COUNT(*) FROM rentals
            WHERE car_id = $1
              AND status IN ('Booked', 'Confirmed', 'Active')
              AND (pickup_datetime < $3 AND dropoff_datetime > $2)`
		errOverlap := tx.Get(&overlapCount, overlapQuery, input.CarID, input.PickupDatetime, input.DropoffDatetime)
		if errOverlap != nil {
			log.Printf("❌ InitiateRentalBooking: Error checking for overlapping rentals for car %d: %v", input.CarID, errOverlap)
			finalErr = fmt.Errorf("failed to verify car availability: %w", errOverlap)
			return models.Rental{}, finalErr
		}
		if overlapCount > 0 {
			log.Printf("❌ InitiateRentalBooking: Car %d is not available due to %d overlapping booking(s) for selected dates despite Availability=true.", input.CarID, overlapCount)
			_, errFixAvail := tx.Exec("UPDATE cars SET availability = false WHERE id = $1", input.CarID)
			if errFixAvail != nil {
				log.Printf("⚠️ Failed to update availability for car %d: %v", input.CarID, errFixAvail)
			}
			finalErr = ErrCarNotAvailable
			return models.Rental{}, finalErr
		}
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
	// --- FIX: Join with cars table to get car details ---
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
	// --- End FIX ---
	err := config.DB.Get(&rental, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Rental{}, ErrRentalNotFound
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
		return rental, ErrForbidden
	}
	return rental, nil
}

func UpdateRentalStatus(rentalID int, newStatus string, employeeID *int) (updatedRental models.Rental, err error) {
	log.Printf("🔄 Service: Attempting to update rental %d status to '%s' (Employee: %v)", rentalID, newStatus, employeeID)

	if rentalID <= 0 {
		err = errors.New("invalid rental ID")
		return
	}
	allowedStatuses := map[string]bool{
		"Confirmed": true, "Active": true, "Returned": true, "Cancelled": true, "Failed": true,
	}
	if !allowedStatuses[newStatus] {
		err = fmt.Errorf("invalid target status for UpdateRentalStatus function: %s", newStatus)
		return
	}

	var tx *sqlx.Tx
	tx, err = config.DB.Beginx()
	if err != nil {
		log.Println("❌ Tx Begin Error:", err)
		err = fmt.Errorf("db transaction error: %w", err)
		return
	}

	defer func() {
		if p := recover(); p != nil {
			log.Println("🔥 Panic during status update, rolling back:", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Printf("❌ Rolling back status update tx due to error: %v", err)
			_ = tx.Rollback()
		} else {
			log.Println("⏳ Committing status update transaction...")
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Println("❌ Error committing status update tx:", commitErr)
				err = fmt.Errorf("commit error: %w", commitErr)
			} else {
				log.Println("✅ Status update tx committed.")
			}
		}
	}()

	var currentStatus string
	var carID int
	query := "SELECT status, car_id FROM rentals WHERE id=$1 FOR UPDATE"
	err = tx.QueryRowx(query, rentalID).Scan(&currentStatus, &carID)
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
		if newStatus == "Cancelled" && employeeID != nil {
			validTransition = true
		}
	case "Booked":
		validTransition = (newStatus == "Confirmed" || newStatus == "Cancelled" || newStatus == "Failed")
	case "Confirmed":
		validTransition = (newStatus == "Active" || newStatus == "Cancelled")
	case "Active":
		validTransition = (newStatus == "Returned")
		if newStatus == "Cancelled" && employeeID != nil {
			validTransition = true
		}
	case "Pending Verification":
		validTransition = (newStatus == "Confirmed" || newStatus == "Failed" || newStatus == "Cancelled")
		log.Printf("⚠️ Warning: Updating rental from 'Pending Verification' status.")
	case "Returned", "Cancelled", "Failed":
		validTransition = false
		log.Printf("ℹ️ Rental %d is already in a terminal state '%s'. Cannot transition to '%s'.", rentalID, currentStatus, newStatus)
	default:
		validTransition = false
		log.Printf("⚠️ Unknown current status '%s' for rental %d", currentStatus, rentalID)
	}

	if !validTransition {
		err = fmt.Errorf("invalid status transition from '%s' to '%s'", currentStatus, newStatus)
		err = ErrInvalidState
		return
	}

	log.Printf("Updating rental %d status from %s to %s", rentalID, currentStatus, newStatus)
	updateQuery := "UPDATE rentals SET status=$1, updated_at = NOW() WHERE id=$2"
	result, execErr := tx.Exec(updateQuery, newStatus, rentalID)
	if execErr != nil {
		err = fmt.Errorf("db error updating rental status: %w", execErr)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		err = ErrRentalNotFound
		return
	}

	updateCarAvailability := false
	newCarAvailability := false
	if newStatus == "Returned" || newStatus == "Cancelled" || newStatus == "Failed" {
		if currentStatus == "Booked" || currentStatus == "Confirmed" || currentStatus == "Active" {
			updateCarAvailability = true
			newCarAvailability = true
		}
	}

	if updateCarAvailability {
		log.Printf("Updating availability for car %d to %t due to rental status change to %s", carID, newCarAvailability, newStatus)
		carUpdateQuery := "UPDATE cars SET availability=$1, updated_at = NOW() WHERE id=$2"
		_, carUpdateErr := tx.Exec(carUpdateQuery, newCarAvailability, carID)
		if carUpdateErr != nil {
			log.Printf("❌ UpdateRentalStatus: Failed to update car availability: %v", carUpdateErr)
			err = fmt.Errorf("failed to update car availability: %w", carUpdateErr)
			return
		} else {
			log.Printf("✅ Marked car %d availability as %t.", carID, newCarAvailability)
		}
	} else {
		log.Printf("ℹ️ No car availability update needed for status change from '%s' to '%s'", currentStatus, newStatus)
	}

	fetchRental, fetchErr := GetRentalByID(rentalID)
	if fetchErr != nil {
		log.Printf("⚠️ Could not fetch updated rental %d after status change: %v", rentalID, fetchErr)
		if err == nil {
			err = fmt.Errorf("status update successful but failed to fetch result: %w", fetchErr)
		}
	} else if err == nil {
		updatedRental = fetchRental
	}

	return
}

func GetRentals() ([]models.Rental, error) {
	var rentals []models.Rental
	log.Println("🔍 Service: Fetching all rentals...")
	// --- FIX: Join with cars table to get car details for admin view too ---
	query := `
		SELECT
			r.id, r.customer_id, r.car_id, r.booking_date,
			r.pickup_datetime, r.dropoff_datetime, r.pickup_location,
			r.status, r.created_at, r.updated_at,
			c.brand AS "car.brand",
			c.model AS "car.model"
		FROM rentals r
		JOIN cars c ON r.car_id = c.id
		ORDER BY r.id ASC`
	// --- End FIX ---
	err := config.DB.Select(&rentals, query)
	if err != nil {
		log.Println("❌ Service: Error fetching rentals:", err)
		return nil, fmt.Errorf("failed to fetch rentals: %w", err)
	}
	log.Printf("✅ Service: Fetched %d rentals successfully", len(rentals))
	return rentals, nil
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
		return ErrRentalNotFound
	}
	log.Println("✅ Service: Rental deleted successfully!")
	return nil
}

func GetRentalsByCustomerID(customerID int) ([]models.Rental, error) {
	var rentals []models.Rental
	log.Println("🔍 Service: Fetching rentals for customer ID:", customerID)
	if customerID <= 0 {
		return nil, errors.New("invalid customer ID")
	}
	// --- FIX: Updated Query with JOIN and Aliases ---
	query := `
		SELECT
			r.id, r.customer_id, r.car_id, r.booking_date,
			r.pickup_datetime, r.dropoff_datetime, r.pickup_location,
			r.status, r.created_at, r.updated_at,
			c.brand AS "car.brand",
			c.model AS "car.model"
		FROM rentals r
		JOIN cars c ON r.car_id = c.id
		WHERE r.customer_id=$1
		ORDER BY r.pickup_datetime DESC`
	// --- End FIX ---
	err := config.DB.Select(&rentals, query, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("ℹ️ Service: No rentals found for customer %d", customerID)
			return []models.Rental{}, nil
		}
		log.Printf("❌ Service: Error fetching rentals for customer %d: %v", customerID, err)
		return nil, fmt.Errorf("failed to fetch rentals for customer %d: %w", customerID, err)
	}
	log.Printf("✅ Service: Fetched %d rentals for customer %d successfully", len(rentals), customerID)
	return rentals, nil
}

func CancelCustomerRental(rentalID int, customerID int) error {
	log.Printf("Service: Customer %d attempting to cancel rental %d", customerID, rentalID)
	if rentalID <= 0 || customerID <= 0 {
		return errors.New("invalid rental or customer ID")
	}
	rental, err := checkRentalOwnership(rentalID, customerID)
	if err != nil {
		return err
	}

	if !(rental.Status == "Pending" || rental.Status == "Booked" || rental.Status == "Confirmed" || rental.Status == "Pending Verification") {
		return fmt.Errorf("cannot cancel rental with status '%s': %w", rental.Status, ErrInvalidState)
	}

	_, err = UpdateRentalStatus(rentalID, "Cancelled", nil)
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
	if hours <= 0 {
		log.Printf("⚠️ CalculateRentalCost: Duration is zero or negative for rental %d.", rentalID)
		return models.Payment{RentalID: rentalID, Amount: 0, PaymentStatus: "Pending"}, nil
	}

	days := duration.Hours() / 24.0
	rentalDays := int(days)
	if days > float64(rentalDays)+1e-9 {
		rentalDays++
	}
	if rentalDays == 0 && hours > 0 {
		rentalDays = 1
	}
	if rentalDays < 0 {
		rentalDays = 0
	}

	cost := float64(rentalDays) * rentalData.Price
	vatRate := 0.07
	vat := cost * vatRate
	totalCost := cost + vat
	log.Printf("✅ Calculated cost for rental %d (%d days): %.2f (Base: %.2f, VAT: %.2f)", rentalID, rentalDays, totalCost, cost, vat)
	return models.Payment{RentalID: rentalID, Amount: totalCost, PaymentStatus: "Pending"}, nil
}
