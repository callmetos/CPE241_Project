package services

import (
	"car-rental-management/config"
	"car-rental-management/models"
	"fmt"
	"log"
)

// GetDashboardData retrieves rental statistics
func GetDashboardData() (models.DashboardData, error) {
	var dashboard models.DashboardData
	log.Println("🔍 Fetching dashboard data...")

	// SQL query to retrieve the dashboard statistics
	query := `
		SELECT 
			(SELECT COUNT(*) FROM rentals) AS total_rentals,
			(SELECT COALESCE(SUM(amount), 0) FROM payments WHERE payment_status='Paid') AS total_revenue,
			(SELECT COUNT(*) FROM customers) AS total_customers
	`

	// Log the query for debugging purposes
	log.Println("Executing query:", query)

	// Attempt to fetch the data
	err := config.DB.Get(&dashboard, query)

	// Log the error in case of failure
	if err != nil {
		log.Printf("❌ Error fetching dashboard data: %v", err) // Log the exact error
		return dashboard, fmt.Errorf("error fetching dashboard data: %v", err)
	}

	log.Println("✅ Dashboard data fetched successfully!")
	return dashboard, nil
}
