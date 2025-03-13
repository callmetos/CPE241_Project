package services

import (
	"car-rental-management/config"
	"car-rental-management/models"
	"log"
)

// Get all users
func GetUsers() ([]models.User, error) {
	var users []models.User
	log.Println("🔍 Fetching users...")

	// Fetch only required fields instead of SELECT *
	err := config.DB.Select(&users, "SELECT id, name, email, role FROM employees")
	if err != nil {
		log.Println("❌ Error fetching users:", err)
		return nil, err
	}

	log.Println("✅ Users fetched successfully! Count:", len(users))
	return users, nil
}
