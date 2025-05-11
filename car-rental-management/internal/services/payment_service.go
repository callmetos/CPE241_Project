package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models" // Ensure models is imported
	"database/sql"
	"errors"
	"fmt"
	"log"

	// "math" // Not directly used here, but CalculateRentalCost in rental_service uses it
	"time"

	"github.com/jmoiron/sqlx"
)

// Note: Error variables like ErrRentalNotFound, ErrForbidden, ErrInvalidState
// are now defined in rental_service.go and are accessible within the 'services' package.

func ProcessPayment(rentalID int, employeeID int, input models.RecordPaymentInput) (models.Payment, error) {
	log.Printf("Service: Processing manual payment record for rental %d by employee %d", rentalID, employeeID)

	if rentalID <= 0 {
		return models.Payment{}, errors.New("invalid rental ID")
	}
	if employeeID <= 0 {
		return models.Payment{}, errors.New("invalid employee ID")
	}

	expectedPaymentData, errCalc := CalculateRentalCost(rentalID) // Call from rental_service
	if errCalc != nil {
		log.Printf("⚠️ ProcessPayment: Could not calculate expected cost for rental %d: %v. Proceeding with input amount.", rentalID, errCalc)
	} else if expectedPaymentData.Amount != input.Amount {
		log.Printf("⚠️ ProcessPayment: Recorded amount %.2f differs from calculated/expected cost %.2f for rental %d", input.Amount, expectedPaymentData.Amount, rentalID)
	}

	payment := models.Payment{
		RentalID:             rentalID,
		Amount:               input.Amount,
		PaymentStatus:        input.PaymentStatus,
		PaymentMethod:        &input.PaymentMethod,
		RecordedByEmployeeID: &employeeID,
		TransactionID:        input.TransactionID,
		PaymentDate:          time.Now(),
		SlipURL:              nil,
	}

	query := `INSERT INTO payments (rental_id, amount, payment_status, payment_method, recorded_by_employee_id, transaction_id, payment_date, slip_url)
			  VALUES (:rental_id, :amount, :payment_status, :payment_method, :recorded_by_employee_id, :transaction_id, :payment_date, :slip_url)
			  RETURNING id, created_at, updated_at`

	tx, errTx := config.DB.Beginx()
	if errTx != nil {
		log.Println("❌ ProcessPayment: Failed to begin transaction:", errTx)
		return models.Payment{}, fmt.Errorf("database transaction error: %w", errTx)
	}

	var finalErr error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if finalErr != nil {
			log.Printf("❌ Rolling back ProcessPayment tx due to error: %v", finalErr)
			_ = tx.Rollback()
		} else {
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Println("❌ Error committing ProcessPayment tx:", commitErr)
				finalErr = fmt.Errorf("commit error: %w", commitErr)
			} else {
				log.Println("✅ ProcessPayment tx committed.")
			}
		}
	}()

	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		log.Println("❌ ProcessPayment: Error preparing insert query:", err)
		finalErr = fmt.Errorf("failed to prepare payment record: %w", err)
		return models.Payment{}, finalErr
	}
	defer stmt.Close()

	err = stmt.QueryRowx(&payment).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		log.Println("❌ ProcessPayment: Error recording payment:", err)
		finalErr = fmt.Errorf("failed to record payment: %w", err)
		return models.Payment{}, finalErr
	}

	if payment.PaymentStatus == "Paid" {
		log.Printf("ℹ️ ProcessPayment: Payment %d recorded as Paid. Attempting to update Rental %d status and car availability.", payment.ID, rentalID)
		// Call UpdateRentalStatus from rental_service, passing the current transaction
		_, updateErr := UpdateRentalStatus(tx, rentalID, "Confirmed", &employeeID)
		if updateErr != nil {
			log.Printf("⚠️ ProcessPayment: Failed to update rental status to Confirmed via UpdateRentalStatus: %v", updateErr)
			finalErr = fmt.Errorf("payment recorded, but failed to update rental/car status: %w", updateErr)
			// The defer func will handle rollback
		} else {
			log.Printf("✅ ProcessPayment: Rental %d status and car availability handled by UpdateRentalStatus.", rentalID)
		}
	}

	log.Printf("✅ Service: Manual payment recorded successfully with ID: %d", payment.ID)
	if finalErr != nil {
		return models.Payment{}, finalErr
	}
	return payment, nil
}

func ProcessSlipUpload(rentalID int, customerID int, slipFilePathOrURL string) (err error) {
	log.Printf("Service: Processing slip upload for rental %d by customer %d. Slip location: %s", rentalID, customerID, slipFilePathOrURL)

	var rentalToCheck struct {
		ID         int    `db:"id"`
		CustomerID int    `db:"customer_id"`
		CarID      int    `db:"car_id"`
		Status     string `db:"status"`
	}
	// This initial check can be outside a transaction
	err = config.DB.Get(&rentalToCheck, "SELECT id, customer_id, car_id, status FROM rentals WHERE id=$1", rentalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ ProcessSlipUpload: Rental %d not found.", rentalID)
			return fmt.Errorf("rental not found: %w", ErrRentalNotFound)
		}
		log.Printf("❌ ProcessSlipUpload: Error fetching rental %d for validation: %v", rentalID, err)
		return fmt.Errorf("database error checking rental: %w", err)
	}

	if rentalToCheck.CustomerID != customerID {
		log.Printf("🚫 ProcessSlipUpload: Permission denied. Customer %d does not own Rental %d.", customerID, rentalID)
		return fmt.Errorf("permission denied: %w", ErrForbidden)
	}

	if rentalToCheck.Status != "Pending" { // Only allow slip upload for "Pending" rentals
		log.Printf("❌ ProcessSlipUpload: Cannot upload slip for rental %d with status '%s', expected 'Pending'", rentalID, rentalToCheck.Status)
		return fmt.Errorf("cannot upload slip for rental with status '%s': %w", rentalToCheck.Status, ErrInvalidState)
	}

	var tx *sqlx.Tx
	tx, errTx := config.DB.Beginx() // Start transaction for payment and rental update
	if errTx != nil {
		log.Printf("❌ ProcessSlipUpload: Failed to begin transaction: %v", errTx)
		return fmt.Errorf("database transaction error: %w", errTx)
	}
	defer func() {
		if p := recover(); p != nil {
			log.Println("🔥 Panic during slip processing, rolling back:", p)
			_ = tx.Rollback()
			panic(p) // Re-panic after rollback
		} else if err != nil {
			log.Printf("❌ Rolling back slip processing tx due to error: %v", err)
			_ = tx.Rollback()
		} else {
			log.Println("⏳ Committing slip processing transaction...")
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Println("❌ Error committing slip processing tx:", commitErr)
				err = fmt.Errorf("commit error: %w", commitErr) // Set outer err so it's returned
			} else {
				log.Println("✅ Slip processing tx committed.")
			}
		}
	}() // This defer will execute when the function returns, with 'err' having its final value

	var paymentID int64
	var currentPaymentStatus string
	// Lock the payment row if it exists to prevent concurrent updates
	paymentQuery := "SELECT id, payment_status FROM payments WHERE rental_id = $1 ORDER BY created_at DESC LIMIT 1 FOR UPDATE"
	dbErr := tx.QueryRowx(paymentQuery, rentalID).Scan(&paymentID, &currentPaymentStatus)

	paymentMethod := "Bank Transfer" // Default for slip upload
	newPaymentStatus := "Pending Verification"
	paymentDate := time.Now()

	if dbErr != nil {
		if errors.Is(dbErr, sql.ErrNoRows) {
			log.Printf("ℹ️ ProcessSlipUpload: No existing payment found for rental %d. Creating new payment record.", rentalID)
			calculatedPaymentData, calcErr := CalculateRentalCost(rentalID)
			if calcErr != nil {
				err = fmt.Errorf("failed to determine payment amount: %w", calcErr)
				return // Defer will rollback
			}
			if calculatedPaymentData.Amount <= 0 {
				err = errors.New("calculated payment amount is invalid or zero")
				return // Defer will rollback
			}
			insertQuery := `INSERT INTO payments (rental_id, amount, payment_status, payment_method, slip_url, payment_date)
                            VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
			err = tx.QueryRowx(insertQuery,
				rentalID, calculatedPaymentData.Amount, newPaymentStatus,
				paymentMethod, slipFilePathOrURL, paymentDate,
			).Scan(&paymentID)
			if err != nil {
				log.Printf("❌ ProcessSlipUpload: Error inserting new payment record: %v", err)
				err = fmt.Errorf("database error creating payment record: %w", err) // Set outer err
				return                                                              // Defer will rollback
			}
			log.Printf("✅ ProcessSlipUpload: New payment record created (ID: %d) with status '%s'", paymentID, newPaymentStatus)
		} else {
			log.Printf("❌ ProcessSlipUpload: Error querying existing payment for rental %d: %v", rentalID, dbErr)
			err = fmt.Errorf("database error checking payment: %w", dbErr) // Set outer err
			return                                                         // Defer will rollback
		}
	} else {
		log.Printf("ℹ️ ProcessSlipUpload: Found existing payment record (ID: %d, Status: %s) for rental %d. Updating.", paymentID, currentPaymentStatus, rentalID)
		// Allow updating if it's Pending (first attempt) or Failed (reattempt)
		if currentPaymentStatus != "Pending" && currentPaymentStatus != "Failed" {
			log.Printf("❌ ProcessSlipUpload: Cannot update payment %d with status '%s' via slip upload.", paymentID, currentPaymentStatus)
			err = fmt.Errorf("cannot re-upload slip for payment in status '%s': %w", currentPaymentStatus, ErrInvalidState) // Set outer err
			return                                                                                                          // Defer will rollback
		}
		updateQuery := `UPDATE payments SET payment_status = $1, slip_url = $2, payment_method = $3, payment_date = $4, updated_at = NOW()
                        WHERE id = $5`
		result, updateErr := tx.Exec(updateQuery, newPaymentStatus, slipFilePathOrURL, paymentMethod, paymentDate, paymentID)
		if updateErr != nil {
			log.Printf("❌ ProcessSlipUpload: Error updating payment record %d: %v", paymentID, updateErr)
			err = fmt.Errorf("database error updating payment: %w", updateErr) // Set outer err
			return                                                             // Defer will rollback
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Printf("❌ ProcessSlipUpload: Payment record %d not found during update.", paymentID)
			err = errors.New("payment record not found during update") // Set outer err
			return                                                     // Defer will rollback
		}
		log.Printf("✅ ProcessSlipUpload: Payment record %d updated to status '%s'", paymentID, newPaymentStatus)
	}

	// If payment processing was successful up to this point, try to update rental status
	// The UpdateRentalStatus function is now tx-aware.
	_, errUpdate := UpdateRentalStatus(tx, rentalID, "Booked", nil) // Pass current transaction tx
	if errUpdate != nil {
		log.Printf("❌ ProcessSlipUpload: Failed to update rental status to 'Booked' and car availability: %v", errUpdate)
		err = fmt.Errorf("database error updating rental status/car availability: %w", errUpdate) // Set outer err for rollback
		return                                                                                    // Defer will rollback
	}
	log.Printf("✅ ProcessSlipUpload: Rental %d status updated to 'Booked' and car availability handled.", rentalID)

	return // If err is nil, defer will commit. If err is set by any step, defer will rollback.
}

func VerifyPayment(rentalId int, approved bool, employeeId int) (err error) {
	log.Printf("🔄 Service: Verifying payment for rental %d. Approved: %t, By Employee: %d", rentalId, approved, employeeId)

	if rentalId <= 0 || employeeId <= 0 {
		return errors.New("invalid rental or employee ID")
	}

	tx, errTx := config.DB.Beginx()
	if errTx != nil {
		log.Printf("❌ VerifyPayment: Failed to begin transaction: %v", errTx)
		return fmt.Errorf("database transaction error: %w", errTx)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Printf("❌ Error committing verify payment tx: %v", commitErr)
				err = fmt.Errorf("commit error: %w", commitErr)
			} else {
				log.Println("✅ Verify payment tx committed.")
			}
		}
	}()

	var fetchedData struct {
		PaymentID     int    `db:"id"`
		RentalID      int    `db:"rental_id"`
		PaymentStatus string `db:"payment_status"`
		RentalStatus  string `db:"rental_status"`
	}
	paymentQuery := `
		SELECT p.id, p.rental_id, p.payment_status, r.status AS rental_status
		FROM payments p
		JOIN rentals r ON p.rental_id = r.id
		WHERE p.rental_id = $1 AND p.payment_status = 'Pending Verification'
		ORDER BY p.created_at DESC LIMIT 1 FOR UPDATE OF p, r`

	err = tx.QueryRowx(paymentQuery, rentalId).StructScan(&fetchedData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ VerifyPayment: No payment pending verification found for rental %d.", rentalId)
			err = errors.New("payment not found or not pending verification")
		} else {
			log.Printf("❌ VerifyPayment: Error querying payment for rental %d: %v", rentalId, err)
			err = fmt.Errorf("database error querying payment: %w", err)
		}
		return // Defer will rollback
	}

	paymentID := fetchedData.PaymentID
	currentRentalStatus := fetchedData.RentalStatus
	var newPaymentStatus string
	var newRentalStatusForUpdate string

	if approved {
		newPaymentStatus = "Paid"
		// Rental must be in a state that allows confirmation
		if currentRentalStatus == "Booked" || currentRentalStatus == "Pending Verification" {
			newRentalStatusForUpdate = "Confirmed"
		} else {
			err = fmt.Errorf("cannot approve rental in '%s' state, expected 'Booked' or 'Pending Verification': %w", currentRentalStatus, ErrInvalidState)
			return // Defer will rollback
		}
		log.Printf("✅ Approving payment for rental %d (Payment ID: %d)", rentalId, paymentID)
	} else { // Rejected
		newPaymentStatus = "Failed" // Or "Rejected" if you have such status
		// Rental must be in a state that allows cancellation/failure due to payment
		if currentRentalStatus == "Booked" || currentRentalStatus == "Pending Verification" {
			newRentalStatusForUpdate = "Cancelled" // Or "FailedPayment" status for rental
		} else {
			err = fmt.Errorf("cannot reject rental in '%s' state, expected 'Booked' or 'Pending Verification': %w", currentRentalStatus, ErrInvalidState)
			return // Defer will rollback
		}
		log.Printf("❌ Rejecting payment for rental %d (Payment ID: %d)", rentalId, paymentID)
	}

	updatePaymentQuery := `UPDATE payments SET payment_status = $1, recorded_by_employee_id = $2, updated_at = NOW() WHERE id = $3`
	_, err = tx.Exec(updatePaymentQuery, newPaymentStatus, employeeId, paymentID)
	if err != nil {
		err = fmt.Errorf("database error updating payment: %w", err)
		return // Defer will rollback
	}
	log.Printf("✅ Payment %d status updated to '%s'", paymentID, newPaymentStatus)

	// Call UpdateRentalStatus using the current transaction 'tx'
	_, err = UpdateRentalStatus(tx, rentalId, newRentalStatusForUpdate, &employeeId)
	if err != nil {
		// err from UpdateRentalStatus will be propagated and cause rollback
		log.Printf("❌ VerifyPayment: Error updating rental status via UpdateRentalStatus: %v", err)
		// No need to set err again, it's already set by UpdateRentalStatus
		return // Defer will rollback
	}
	log.Printf("✅ Rental %d status updated to '%s' and car availability handled via UpdateRentalStatus.", rentalId, newRentalStatusForUpdate)

	return // If err is nil, commit will happen. Otherwise, rollback.
}

func GetPayments() ([]models.Payment, error) {
	var payments []models.Payment
	query := "SELECT id, rental_id, amount, payment_date, payment_status, payment_method, recorded_by_employee_id, transaction_id, slip_url, created_at, updated_at FROM payments ORDER BY id ASC"
	err := config.DB.Select(&payments, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payments: %w", err)
	}
	return payments, nil
}

func GetPaymentsByRentalID(rentalID int) ([]models.Payment, error) {
	var payments []models.Payment
	if rentalID <= 0 {
		return nil, errors.New("invalid rental ID")
	}
	query := "SELECT id, rental_id, amount, payment_date, payment_status, payment_method, recorded_by_employee_id, transaction_id, slip_url, created_at, updated_at FROM payments WHERE rental_id=$1 ORDER BY id ASC"
	err := config.DB.Select(&payments, query, rentalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.Payment{}, nil
		}
		return nil, fmt.Errorf("failed to fetch payments for rental %d: %w", rentalID, err)
	}
	return payments, nil
}

func GetPaymentStatus(paymentID int) (models.Payment, error) {
	if paymentID <= 0 {
		return models.Payment{}, errors.New("invalid payment ID")
	}
	var payment models.Payment
	query := "SELECT id, rental_id, payment_status FROM payments WHERE id=$1"
	err := config.DB.Get(&payment, query, paymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Payment{}, errors.New("payment not found") // Direct error, not wrapped from rental_service
		}
		return models.Payment{}, fmt.Errorf("failed to fetch payment status: %w", err)
	}
	return payment, nil
}

type RentalPendingVerification struct {
	RentalID        int       `db:"rental_id" json:"rental_id"`
	CustomerID      int       `db:"customer_id" json:"customer_id"`
	CustomerName    string    `db:"customer_name" json:"customer_name"`
	CarID           int       `db:"car_id" json:"car_id"`
	CarBrand        string    `db:"car_brand" json:"car_brand"`
	CarModel        string    `db:"car_model" json:"car_model"`
	PaymentID       int       `db:"payment_id" json:"payment_id"`
	PaymentAmount   float64   `db:"payment_amount" json:"payment_amount"`
	SlipURL         *string   `db:"slip_url" json:"slip_url"`
	PaymentDate     time.Time `db:"payment_date" json:"payment_date"`
	PickupDatetime  time.Time `db:"pickup_datetime" json:"pickup_datetime"`
	DropoffDatetime time.Time `db:"dropoff_datetime" json:"dropoff_datetime"`
}

func GetRentalsPendingVerification() ([]RentalPendingVerification, error) {
	var rentals []RentalPendingVerification
	query := `
		SELECT
			r.id AS rental_id, r.customer_id, cust.name AS customer_name,
			r.car_id, ca.brand AS car_brand, ca.model AS car_model,
			p.id AS payment_id, p.amount AS payment_amount, p.slip_url, p.payment_date,
			r.pickup_datetime, r.dropoff_datetime
		FROM rentals r
		JOIN payments p ON r.id = p.rental_id
		JOIN customers cust ON r.customer_id = cust.id
		JOIN cars ca ON r.car_id = ca.id
		WHERE p.payment_status = 'Pending Verification'
		ORDER BY p.payment_date ASC
	`
	err := config.DB.Select(&rentals, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []RentalPendingVerification{}, nil
		}
		return nil, fmt.Errorf("failed to fetch rentals pending verification: %w", err)
	}
	return rentals, nil
}
