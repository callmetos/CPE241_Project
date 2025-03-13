package services

import (
	"car-rental-management/config"
	"car-rental-management/models"
	"log"
)

// Process a payment
func ProcessPayment(payment models.Payment) error {
	log.Println("🔍 Attempting to process payment:", payment)

	// Insert payment into the database
	_, err := config.DB.NamedExec("INSERT INTO payments (rental_id, amount, payment_status) VALUES (:rental_id, :amount, :payment_status)", payment)

	if err != nil {
		log.Println("❌ Error processing payment:", err)
		return err
	}

	log.Println("✅ Payment processed successfully!")
	return nil
}

// Get all payments
func GetPayments() ([]models.Payment, error) {
	var payments []models.Payment

	log.Println("🔍 Fetching all payments from the database...")

	// Execute the database query
	err := config.DB.Select(&payments, "SELECT * FROM payments")

	if err != nil {
		log.Println("❌ Error fetching payments:", err)
		return nil, err
	}

	log.Println("✅ Payments fetched successfully:", payments)
	return payments, nil
}
