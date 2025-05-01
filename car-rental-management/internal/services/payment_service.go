// internal/services/payment_service.go
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

// *** FIX: Remove duplicate error declarations (defined in rental_service.go) ***
/*
var (
	ErrRentalNotFound  = errors.New("rental not found")
	ErrCarNotFound     = errors.New("car not found")
	ErrCarNotAvailable = errors.New("car is not available for the selected dates")
	ErrInvalidDates    = errors.New("invalid pickup/dropoff dates")
	ErrForbidden       = errors.New("permission denied")
	ErrInvalidState    = errors.New("invalid operation for current rental/payment state")
)
*/

// --- *** ฟังก์ชัน Service ใหม่: ProcessSlipUpload *** ---
// Processes the slip upload: checks ownership, updates payment status, saves slip URL, and updates rental status.
func ProcessSlipUpload(rentalID int, customerID int, slipFilePathOrURL string) (err error) { // Use named return for easier defer error handling
	log.Printf("Service: Processing slip upload for rental %d by customer %d. Slip location: %s", rentalID, customerID, slipFilePathOrURL)

	// 1. Check Ownership & Rental Status
	var rentalToCheck models.Rental
	err = config.DB.Get(&rentalToCheck, "SELECT id, customer_id, status FROM rentals WHERE id=$1", rentalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ ProcessSlipUpload: Rental %d not found.", rentalID)
			return fmt.Errorf("rental not found") // Use standard error or defined ErrRentalNotFound
		}
		log.Printf("❌ ProcessSlipUpload: Error fetching rental %d for validation: %v", rentalID, err)
		return fmt.Errorf("database error checking rental: %w", err)
	}

	// Check ownership
	if rentalToCheck.CustomerID != customerID {
		log.Printf("🚫 ProcessSlipUpload: Permission denied. Customer %d does not own Rental %d.", customerID, rentalID)
		return fmt.Errorf("permission denied") // Use standard error or defined ErrForbidden
	}

	// Check current rental status (should be Pending)
	if rentalToCheck.Status != "Pending" {
		log.Printf("❌ ProcessSlipUpload: Cannot upload slip for rental %d with status '%s'", rentalID, rentalToCheck.Status)
		return fmt.Errorf("invalid operation for current rental state") // Use standard error or defined ErrInvalidState
	}

	// 3. Start Transaction
	var tx *sqlx.Tx // Declare tx variable
	tx, errTx := config.DB.Beginx()
	if errTx != nil {
		log.Printf("❌ ProcessSlipUpload: Failed to begin transaction: %v", errTx)
		return fmt.Errorf("database transaction error: %w", errTx)
	}
	// Defer rollback/commit logic using named return 'err'
	defer func() {
		if p := recover(); p != nil {
			log.Println("🔥 Panic during slip processing, rolling back:", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil { // Check the named return error 'err'
			log.Printf("❌ Rolling back slip processing tx due to error: %v", err)
			_ = tx.Rollback()
		} else {
			log.Println("⏳ Committing slip processing transaction...")
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Println("❌ Error committing slip processing tx:", commitErr)
				err = fmt.Errorf("commit error: %w", commitErr) // Assign commit error to 'err'
			} else {
				log.Println("✅ Slip processing tx committed.")
			}
		}
	}()

	// 4. Find or Create Payment Record within Transaction
	var paymentID int64
	var currentPaymentStatus string
	paymentQuery := "SELECT id, payment_status FROM payments WHERE rental_id = $1 ORDER BY created_at DESC LIMIT 1 FOR UPDATE" // Lock row
	dbErr := tx.QueryRowx(paymentQuery, rentalID).Scan(&paymentID, &currentPaymentStatus)

	paymentMethod := "Bank Transfer" // Assuming slip upload implies bank transfer
	newPaymentStatus := "Pending Verification"
	paymentDate := time.Now()

	if dbErr != nil {
		if errors.Is(dbErr, sql.ErrNoRows) {
			// --- Create New Payment Record ---
			log.Printf("ℹ️ ProcessSlipUpload: No existing payment found for rental %d. Creating new payment record.", rentalID)
			// *** Call CalculateRentalCost from rental_service (it's in the same package) ***
			calculatedPaymentData, calcErr := CalculateRentalCost(rentalID)
			if calcErr != nil {
				err = fmt.Errorf("failed to determine payment amount: %w", calcErr)
				return err // Trigger rollback
			}
			if calculatedPaymentData.Amount <= 0 {
				err = errors.New("calculated payment amount is invalid")
				return err // Trigger rollback
			}

			insertQuery := `INSERT INTO payments (rental_id, amount, payment_status, payment_method, slip_url, payment_date)
                            VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
			err = tx.QueryRowx(insertQuery,
				rentalID, calculatedPaymentData.Amount, newPaymentStatus,
				paymentMethod, slipFilePathOrURL, paymentDate,
			).Scan(&paymentID)

			if err != nil {
				log.Printf("❌ ProcessSlipUpload: Error inserting new payment record: %v", err)
				err = fmt.Errorf("database error creating payment record: %w", err)
				return err // Trigger rollback
			}
			log.Printf("✅ ProcessSlipUpload: New payment record created (ID: %d) with status '%s'", paymentID, newPaymentStatus)

		} else {
			// Other error querying payment
			log.Printf("❌ ProcessSlipUpload: Error querying existing payment for rental %d: %v", rentalID, dbErr)
			err = fmt.Errorf("database error checking payment: %w", dbErr)
			return err // Trigger rollback
		}
	} else {
		// --- Update Existing Payment Record ---
		log.Printf("ℹ️ ProcessSlipUpload: Found existing payment record (ID: %d, Status: %s) for rental %d. Updating.", paymentID, currentPaymentStatus, rentalID)
		if currentPaymentStatus != "Pending" && currentPaymentStatus != "Failed" { // Allow retry from Failed?
			log.Printf("❌ ProcessSlipUpload: Cannot update payment %d with status '%s'", paymentID, currentPaymentStatus)
			err = fmt.Errorf("invalid operation for current payment state") // Use standard error or defined ErrInvalidState
			return err                                                      // Trigger rollback
		}

		updateQuery := `UPDATE payments SET payment_status = $1, slip_url = $2, payment_method = $3, payment_date = $4, updated_at = NOW()
                        WHERE id = $5`
		result, updateErr := tx.Exec(updateQuery, newPaymentStatus, slipFilePathOrURL, paymentMethod, paymentDate, paymentID)
		if updateErr != nil {
			log.Printf("❌ ProcessSlipUpload: Error updating payment record %d: %v", paymentID, updateErr)
			err = fmt.Errorf("database error updating payment: %w", updateErr)
			return err // Trigger rollback
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Printf("❌ ProcessSlipUpload: Payment record %d not found during update.", paymentID)
			err = errors.New("payment record not found during update")
			return err // Trigger rollback
		}
		log.Printf("✅ ProcessSlipUpload: Payment record %d updated to status '%s'", paymentID, newPaymentStatus)
	}

	// *** FIX 1: Update Rental Status to "Booked" ***
	if err == nil { // Check if payment processing above was successful
		log.Printf("ℹ️ ProcessSlipUpload: Attempting to update rental %d status to 'Booked'", rentalID)
		updateRentalQuery := `UPDATE rentals SET status = 'Booked', updated_at = NOW() WHERE id = $1 AND status = 'Pending'`
		resultRentalUpdate, errUpdateRental := tx.Exec(updateRentalQuery, rentalID)

		if errUpdateRental != nil {
			log.Printf("❌ ProcessSlipUpload: Failed to update rental status to 'Booked': %v", errUpdateRental)
			err = fmt.Errorf("database error updating rental status: %w", errUpdateRental)
		} else {
			rowsAffectedRental, _ := resultRentalUpdate.RowsAffected()
			if rowsAffectedRental == 0 {
				log.Printf("⚠️ ProcessSlipUpload: Rental %d status not updated to 'Booked' (rows affected: 0, maybe status was not 'Pending'?)", rentalID)
				err = fmt.Errorf("failed to update rental status, rental ID %d not found or status was not 'Pending'", rentalID)
			} else {
				log.Printf("✅ ProcessSlipUpload: Rental %d status updated to 'Booked'.", rentalID)
			}
		}
	}
	// *** End: Added section ***

	return err // Return the named error variable 'err'
}

// Staff records a payment manually
func ProcessPayment(rentalID int, employeeID int, input models.RecordPaymentInput) (models.Payment, error) {
	log.Printf("Service: Processing manual payment record for rental %d by employee %d", rentalID, employeeID)

	if rentalID <= 0 {
		return models.Payment{}, errors.New("invalid rental ID")
	}
	if employeeID <= 0 {
		return models.Payment{}, errors.New("invalid employee ID")
	}

	// Assume CalculateRentalCost exists and is accessible
	expectedPaymentData, errCalc := CalculateRentalCost(rentalID)
	if errCalc != nil {
		log.Printf("⚠️ ProcessPayment: Could not calculate expected cost for rental %d: %v.", rentalID, errCalc)
	} else if expectedPaymentData.Amount != input.Amount {
		log.Printf("⚠️ ProcessPayment: Recorded amount %.2f differs from calculated cost %.2f for rental %d", input.Amount, expectedPaymentData.Amount, rentalID)
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
			finalErr = tx.Commit()
			if finalErr != nil {
				log.Println("❌ Error committing ProcessPayment tx:", finalErr)
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
		log.Printf("ℹ️ ProcessPayment: Payment %d recorded as Paid. Attempting to update Rental %d status within transaction.", payment.ID, rentalID)
		newRentalStatus := "Confirmed"
		updateRentalQuery := `UPDATE rentals SET status = $1, booking_date = NOW(), updated_at = NOW() WHERE id = $2 AND status != 'Returned' AND status != 'Cancelled'`
		result, updateErr := tx.Exec(updateRentalQuery, newRentalStatus, rentalID)
		if updateErr != nil {
			log.Printf("⚠️ ProcessPayment: Failed to update rental status to %s within transaction: %v", newRentalStatus, updateErr)
			finalErr = fmt.Errorf("failed to update rental status after payment: %w", updateErr)
			return models.Payment{}, finalErr
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("✅ ProcessPayment: Rental %d status updated to %s.", rentalID, newRentalStatus)
		} else {
			log.Printf("⚠️ ProcessPayment: Rental %d status was not updated (rows affected: 0, maybe status was already Returned/Cancelled?).", rentalID)
		}
	}

	log.Printf("✅ Service: Manual payment recorded successfully with ID: %d", payment.ID)
	return payment, finalErr
}

func GetPayments() ([]models.Payment, error) {
	var payments []models.Payment
	log.Println("🔍 Service: Fetching all payments...")
	query := "SELECT id, rental_id, amount, payment_date, payment_status, payment_method, recorded_by_employee_id, transaction_id, slip_url, created_at, updated_at FROM payments ORDER BY id ASC"
	err := config.DB.Select(&payments, query)
	if err != nil {
		log.Println("❌ Service: Error fetching payments:", err)
		return nil, fmt.Errorf("failed to fetch payments: %w", err)
	}
	log.Printf("✅ Service: Fetched %d payments successfully", len(payments))
	return payments, nil
}

func GetPaymentsByRentalID(rentalID int) ([]models.Payment, error) {
	var payments []models.Payment
	log.Println("🔍 Service: Fetching payments for rental ID:", rentalID)
	if rentalID <= 0 {
		return nil, errors.New("invalid rental ID")
	}
	query := "SELECT id, rental_id, amount, payment_date, payment_status, payment_method, recorded_by_employee_id, transaction_id, slip_url, created_at, updated_at FROM payments WHERE rental_id=$1 ORDER BY id ASC"
	err := config.DB.Select(&payments, query, rentalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("ℹ️ Service: No payments found for rental %d", rentalID)
			return []models.Payment{}, nil
		}
		log.Printf("❌ Service: Error fetching payments for rental %d: %v", rentalID, err)
		return nil, fmt.Errorf("failed to fetch payments for rental %d: %w", rentalID, err)
	}
	log.Printf("✅ Service: Payments fetched successfully for rental %d: Count %d", rentalID, len(payments))
	return payments, nil
}

func GetPaymentStatus(paymentID int) (models.Payment, error) {
	log.Printf("Service: Fetching status for payment ID: %d", paymentID)
	if paymentID <= 0 {
		return models.Payment{}, errors.New("invalid payment ID")
	}

	var payment models.Payment
	query := "SELECT id, rental_id, payment_status FROM payments WHERE id=$1"
	err := config.DB.Get(&payment, query, paymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ Service: Payment %d not found.", paymentID)
			return models.Payment{}, errors.New("payment not found")
		}
		log.Printf("❌ Service: Error fetching payment %d status: %v", paymentID, err)
		return models.Payment{}, fmt.Errorf("failed to fetch payment status: %w", err)
	}
	log.Printf("✅ Service: Status for payment %d is '%s'", paymentID, payment.PaymentStatus)
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
	log.Println("🔍 Service: Fetching rentals pending verification...")

	query := `
		SELECT
			r.id AS rental_id,
			r.customer_id,
			c.name AS customer_name,
			r.car_id,
			ca.brand AS car_brand,
			ca.model AS car_model,
			p.id AS payment_id,
			p.amount AS payment_amount,
			p.slip_url,
			p.payment_date,
			r.pickup_datetime,
			r.dropoff_datetime
		FROM rentals r
		JOIN payments p ON r.id = p.rental_id
		JOIN customers c ON r.customer_id = c.id
		JOIN cars ca ON r.car_id = ca.id
		WHERE p.payment_status = 'Pending Verification'
		ORDER BY p.payment_date ASC
	`
	err := config.DB.Select(&rentals, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("ℹ️ Service: No rentals pending verification found.")
			return []RentalPendingVerification{}, nil
		}
		log.Printf("❌ Service: Error fetching rentals pending verification: %v", err)
		return nil, fmt.Errorf("failed to fetch rentals pending verification: %w", err)
	}

	log.Printf("✅ Service: Fetched %d rentals pending verification.", len(rentals))
	return rentals, nil
}

// *** FIX 3 (v3): Refactor VerifyPayment using Queryx and StructScan ***
func VerifyPayment(rentalId int, approved bool, employeeId int) (err error) { // ใช้ named return
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
			log.Printf("❌ Rolling back verify payment tx due to error: %v", err)
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

	// Struct ชั่วคราวเพื่อรับข้อมูล
	var fetchedData struct {
		PaymentID     int    `db:"id"`
		RentalID      int    `db:"rental_id"`
		PaymentStatus string `db:"payment_status"`
		CarID         int    `db:"car_id"`
		RentalStatus  string `db:"rental_status"`
	}

	// Query พร้อม Alias
	paymentQuery := `
		SELECT
			p.id AS id,
			p.rental_id AS rental_id,
			p.payment_status AS payment_status,
			r.car_id AS car_id,
			r.status AS rental_status
		FROM payments p
		JOIN rentals r ON p.rental_id = r.id
		WHERE p.rental_id = $1 AND p.payment_status = 'Pending Verification'
		FOR UPDATE OF p, r`

	// ใช้ Queryx และ StructScan
	rows, errQuery := tx.Queryx(paymentQuery, rentalId)
	if errQuery != nil {
		log.Printf("❌ VerifyPayment: Error executing query for rental %d: %v", rentalId, errQuery)
		err = fmt.Errorf("database error executing query: %w", errQuery)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		errScan := rows.StructScan(&fetchedData)
		if errScan != nil {
			log.Printf("❌ VerifyPayment: Error scanning row into struct for rental %d: %v", rentalId, errScan)
			err = fmt.Errorf("database error scanning data: %w", errScan)
			return err
		}
	} else {
		log.Printf("❌ VerifyPayment: No payment pending verification found for rental %d (rows.Next() was false).", rentalId)
		err = errors.New("payment not found or not pending verification")
		return err
	}

	if rows.Next() {
		log.Printf("⚠️ VerifyPayment: Found multiple payments pending verification for rental %d. This should not happen.", rentalId)
	}

	// ใช้ข้อมูลจาก fetchedData
	paymentID := fetchedData.PaymentID
	carId := fetchedData.CarID
	currentRentalStatus := fetchedData.RentalStatus

	// กำหนดสถานะใหม่
	var newPaymentStatus string
	var newRentalStatus string
	var carShouldBeAvailable bool = false

	if approved {
		newPaymentStatus = "Paid"
		if currentRentalStatus == "Booked" {
			newRentalStatus = "Confirmed"
		} else {
			log.Printf("⚠️ VerifyPayment: Cannot approve rental %d with current status '%s'. Expected 'Booked'.", rentalId, currentRentalStatus)
			err = fmt.Errorf("cannot approve rental in '%s' state", currentRentalStatus)
			return err
		}
		carShouldBeAvailable = false
		log.Printf("✅ Approving payment for rental %d (Payment ID: %d)", rentalId, paymentID)
	} else {
		newPaymentStatus = "Failed"
		if currentRentalStatus == "Booked" {
			newRentalStatus = "Cancelled"
		} else {
			log.Printf("⚠️ VerifyPayment: Cannot reject rental %d with current status '%s'. Expected 'Booked'.", rentalId, currentRentalStatus)
			err = fmt.Errorf("cannot reject rental in '%s' state", currentRentalStatus)
			return err
		}
		carShouldBeAvailable = true
		log.Printf("❌ Rejecting payment for rental %d (Payment ID: %d)", rentalId, paymentID)
	}

	// อัปเดต Payment Status
	updatePaymentQuery := `UPDATE payments SET payment_status = $1, recorded_by_employee_id = $2, updated_at = NOW() WHERE id = $3`
	_, err = tx.Exec(updatePaymentQuery, newPaymentStatus, employeeId, paymentID)
	if err != nil {
		log.Printf("❌ VerifyPayment: Error updating payment status for payment %d: %v", paymentID, err)
		err = fmt.Errorf("database error updating payment: %w", err)
		return err
	}
	log.Printf("✅ Payment %d status updated to '%s'", paymentID, newPaymentStatus)

	// อัปเดต Rental Status
	updateRentalQuery := `UPDATE rentals SET status = $1, updated_at = NOW()`
	if approved {
		updateRentalQuery += ", booking_date = NOW()"
	}
	updateRentalQuery += " WHERE id = $2"
	resultRental, errUpdateRental := tx.Exec(updateRentalQuery, newRentalStatus, rentalId)
	if errUpdateRental != nil {
		log.Printf("❌ VerifyPayment: Error updating rental status for rental %d: %v", rentalId, errUpdateRental)
		err = fmt.Errorf("database error updating rental: %w", errUpdateRental)
		return err
	}
	rowsAffectedRental, _ := resultRental.RowsAffected()
	if rowsAffectedRental == 0 {
		log.Printf("❌ VerifyPayment: Failed to update rental %d status (rows affected 0).", rentalId)
		err = errors.New("failed to update rental status, rental not found during update")
		return err
	}
	log.Printf("✅ Rental %d status updated to '%s'", rentalId, newRentalStatus)

	// อัปเดต Car Availability
	if carShouldBeAvailable {
		updateCarQuery := `UPDATE cars SET availability = true, updated_at = NOW() WHERE id = $1`
		resultCar, errUpdateCar := tx.Exec(updateCarQuery, carId)
		if errUpdateCar != nil {
			log.Printf("❌ VerifyPayment: Error updating car availability for car %d: %v", carId, errUpdateCar)
			err = fmt.Errorf("database error updating car: %w", errUpdateCar)
		} else {
			rowsAffectedCar, _ := resultCar.RowsAffected()
			if rowsAffectedCar == 0 {
				log.Printf("⚠️ VerifyPayment: Failed to update availability for car %d (rows affected 0).", carId)
			} else {
				log.Printf("✅ Car %d availability set to true after rejection/cancellation.", carId)
			}
		}
	}

	return err // Will be nil if everything succeeded
}

// *** FIX: Remove duplicate CalculateRentalCost function ***
/*
func CalculateRentalCost(rentalID int) (models.Payment, error) {
	// ... implementation ...
}
*/
