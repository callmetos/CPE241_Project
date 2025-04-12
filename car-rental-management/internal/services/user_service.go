package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"log"
)

// GetUsers retrieves a list of employees (excluding passwords)
func GetUsers() ([]models.Employee, error) { // Return type Employee for clarity
	var users []models.Employee // Use Employee model directly
	log.Println("🔍 Fetching users (employees)...")

	// Fetch necessary fields (exclude password)
	query := "SELECT id, name, email, role, created_at, updated_at FROM employees ORDER BY name"
	err := config.DB.Select(&users, query)
	if err != nil {
		log.Println("❌ Error fetching users (employees):", err)
		return nil, err // Return the original error
	}

	log.Printf("✅ Users (employees) fetched successfully! Count: %d", len(users))
	return users, nil
}
