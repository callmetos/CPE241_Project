// ในไฟล์ internal/services/payment_service.go
package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"database/sql"
	"errors"
	"log"
	"strings" // Added for error checking maybe
	"time"    // Import time สำหรับ CalculateRentalCost
)

// ProcessPayment records a payment made for a rental (likely by staff)
func ProcessPayment(rentalID int, employeeID int, input models.RecordPaymentInput) (models.Payment, error) {
	log.Printf("Processing payment record for rental %d by employee %d", rentalID, employeeID)

	// Optional: Fetch rental details to verify status ('Returned'?) or calculate expected amount
	expectedAmount, errCalc := CalculateRentalCost(rentalID) // <--- เรียกใช้ CalculateRentalCost
	if errCalc != nil {
		log.Printf("⚠️ Could not calculate expected cost for rental %d: %v. Proceeding with provided amount.", rentalID, errCalc)
		// Decide if this should be a hard error or just a warning
	} else if expectedAmount != input.Amount {
		log.Printf("⚠️ Warning: Recorded payment amount %.2f differs from calculated cost %.2f for rental %d", input.Amount, expectedAmount, rentalID)
	}

	payment := models.Payment{
		RentalID:             rentalID,
		Amount:               input.Amount,
		PaymentStatus:        input.PaymentStatus,
		PaymentMethod:        &input.PaymentMethod, // Use address-of for pointer field
		RecordedByEmployeeID: &employeeID,
		TransactionID:        input.TransactionID,
		// PaymentDate defaults to CURRENT_TIMESTAMP in DB
	}

	query := `INSERT INTO payments (rental_id, amount, payment_status, payment_method, recorded_by_employee_id, transaction_id)
			  VALUES (:rental_id, :amount, :payment_status, :payment_method, :recorded_by_employee_id, :transaction_id)
			  RETURNING id, payment_date, created_at, updated_at` // Return generated fields

	// Use NamedQuery or similar to execute and scan into payment struct
	stmt, err := config.DB.PrepareNamed(query)
	if err != nil {
		log.Println("❌ Error preparing payment insert query:", err)
		return models.Payment{}, errors.New("failed to prepare payment record")
	}
	defer stmt.Close()

	// Scan the returned values into the corresponding fields of the payment struct
	err = stmt.QueryRowx(payment).Scan(&payment.ID, &payment.PaymentDate, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		log.Println("❌ Error recording payment:", err)
		// Check for FK violation (rental_id exists?)
		if strings.Contains(err.Error(), "payments_rental_id_fkey") {
			return models.Payment{}, errors.New("cannot record payment: rental not found")
		}
		return models.Payment{}, errors.New("failed to record payment")
	}

	log.Printf("✅ Payment recorded successfully with ID: %d", payment.ID)
	return payment, nil
}

// GetPayments retrieves all payments (consider adding filtering/pagination)
func GetPayments() ([]models.Payment, error) {
	var payments []models.Payment
	log.Println("🔍 Fetching all payments from the database...")
	query := "SELECT id, rental_id, amount, payment_date, payment_status, payment_method, recorded_by_employee_id, transaction_id, created_at, updated_at FROM payments ORDER BY payment_date DESC"
	err := config.DB.Select(&payments, query)
	if err != nil {
		log.Println("❌ Error fetching payments:", err)
		return nil, errors.New("failed to fetch payments")
	}
	log.Println("✅ Payments fetched successfully:", len(payments))
	return payments, nil
}

// --- ฟังก์ชันที่อาจจะขาดไป ---

// GetPaymentsByRentalID retrieves payments for a specific rental
func GetPaymentsByRentalID(rentalID int) ([]models.Payment, error) {
	var payments []models.Payment
	log.Println("🔍 Fetching payments for rental ID:", rentalID)
	query := "SELECT id, rental_id, amount, payment_date, payment_status, payment_method, recorded_by_employee_id, transaction_id, created_at, updated_at FROM payments WHERE rental_id=$1 ORDER BY payment_date DESC"
	err := config.DB.Select(&payments, query, rentalID)
	if err != nil {
		log.Printf("❌ Error fetching payments for rental %d: %v", rentalID, err)
		return nil, errors.New("failed to fetch payments for rental")
	}
	log.Printf("✅ Payments fetched successfully for rental %d: Count %d", rentalID, len(payments))
	return payments, nil
}

// CalculateRentalCost calculates the cost based on duration and car price
func CalculateRentalCost(rentalID int) (float64, error) {
	log.Println("Calculating cost for rental ID:", rentalID)

	var rentalData struct {
		Pickup  time.Time `db:"pickup_datetime"`
		Dropoff time.Time `db:"dropoff_datetime"`
		Price   float64   `db:"price_per_day"`
		CarID   int       `db:"car_id"` // Included CarID for logging/verification
	}

	// Query to get rental times and car price per day
	query := `SELECT r.pickup_datetime, r.dropoff_datetime, c.price_per_day, r.car_id
               FROM rentals r
               JOIN cars c ON r.car_id = c.id
               WHERE r.id=$1`

	err := config.DB.Get(&rentalData, query, rentalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0.0, errors.New("rental not found for cost calculation")
		}
		log.Printf("❌ Error fetching rental/car data for cost calculation (rental %d): %v", rentalID, err)
		return 0.0, errors.New("failed to get data for cost calculation")
	}

	// Ensure dates are valid (Dropoff is after Pickup)
	if !rentalData.Dropoff.After(rentalData.Pickup) {
		log.Printf("❌ Invalid dates for rental %d: Pickup=%v, Dropoff=%v", rentalID, rentalData.Pickup, rentalData.Dropoff)
		return 0.0, errors.New("invalid rental dates for cost calculation")
	}

	// Calculate duration
	duration := rentalData.Dropoff.Sub(rentalData.Pickup)

	// --- Logic for calculating rental days ---
	// Example: Round up to the nearest full day.
	// Consider if business logic requires hourly rates, different rounding, minimum days etc.
	hours := duration.Hours()
	if hours <= 0 {
		log.Printf("⚠️ Rental %d has zero or negative duration. Setting cost to 0.", rentalID)
		return 0.0, nil // Or return an error? Or minimum charge?
	}

	// Calculate days, rounding up
	days := hours / 24.0
	rentalDays := int(days)
	if float64(rentalDays) < days {
		rentalDays++ // Round up
	}
	// Ensure at least one day is charged if duration > 0?
	if rentalDays == 0 && hours > 0 {
		rentalDays = 1
	}
	// --- End Day Calculation Logic ---

	// Calculate final cost
	cost := float64(rentalDays) * rentalData.Price

	log.Printf("✅ Calculated cost for rental %d: %.2f (%d days * %.2f/day)", rentalID, cost, rentalDays, rentalData.Price)
	return cost, nil
}
