package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models" // <--- ตรวจสอบว่ามี import นี้
	"fmt"
	"log"
)

// GetDashboardData retrieves rental statistics dynamically
func GetDashboardData() (models.DashboardData, error) {
	var dashboard models.DashboardData // <--- ใช้งาน struct
	log.Println("🔍 Fetching dashboard data...")

	query := `
		SELECT
			(SELECT COUNT(*) FROM rentals) AS total_rentals,
			(SELECT COALESCE(SUM(amount), 0) FROM payments WHERE payment_status='Paid') AS total_revenue,
			(SELECT COUNT(*) FROM customers) AS total_customers
	`

	log.Println("Executing dashboard query:", query)

	err := config.DB.Get(&dashboard, query) // <--- ใช้งาน struct
	if err != nil {
		log.Printf("❌ Error fetching dashboard data: %v", err)
		return models.DashboardData{}, fmt.Errorf("error fetching dashboard data: %v", err) // <--- ใช้งาน struct
	}

	log.Println("✅ Dashboard data fetched successfully!")
	return dashboard, nil
}
