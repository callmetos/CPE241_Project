package services

import (
	"car-rental-management/config"
	"car-rental-management/models"
	"log"
)

// GetCustomers retrieves all customers from the database
func GetCustomers() ([]models.Customer, error) {
	var customers []models.Customer

	log.Println("🔍 Attempting to fetch customers from the database...")

	// Execute the database query
	err := config.DB.Select(&customers, "SELECT * FROM customers")

	if err != nil {
		log.Println("❌ Database query error:", err)
		return nil, err
	}

	log.Println("✅ Customers fetched successfully:", customers)
	return customers, nil
}
