package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

func GetDashboardData() (models.DashboardData, error) {
	var dashboard models.DashboardData
	log.Println("🔍 Fetching dashboard data for admin...")

	query := `
		SELECT
			(SELECT COUNT(*) FROM rentals) AS total_rentals,
			(SELECT COALESCE(SUM(amount), 0) FROM payments WHERE payment_status='Paid') AS total_revenue,
			(SELECT COUNT(*) FROM customers) AS total_customers,
			(SELECT COUNT(*) FROM cars) AS total_cars,
			(SELECT COUNT(*) FROM cars WHERE availability = TRUE) AS total_available_cars,
			(SELECT COUNT(*) FROM cars WHERE availability = FALSE) AS unavailable_cars,
			(SELECT COUNT(*) FROM branches) AS total_branches
	`

	log.Println("Executing admin dashboard query:", query)

	err := config.DB.Get(&dashboard, query)
	if err != nil {
		log.Printf("❌ Error fetching admin dashboard data: %v", err)
		return models.DashboardData{}, fmt.Errorf("error fetching admin dashboard data: %v", err)
	}

	log.Println("✅ Admin dashboard data fetched successfully!")
	return dashboard, nil
}

func GetPublicStatsData() (models.PublicStatsData, error) {
	var stats models.PublicStatsData
	log.Println("🔍 Fetching public stats data...")

	query := `
		SELECT
			(SELECT COUNT(*) FROM cars WHERE availability = TRUE) AS total_available_cars,
			(SELECT COUNT(*) FROM branches) AS total_branches
	`
	log.Println("Executing public stats query:", query)
	err := config.DB.Get(&stats, query)
	if err != nil {
		log.Printf("❌ Error fetching public stats data: %v", err)
		return models.PublicStatsData{}, fmt.Errorf("error fetching public stats data: %v", err)
	}
	log.Println("✅ Public stats data fetched successfully!")
	return stats, nil
}

func GetRevenueReport(startDate, endDate time.Time) ([]models.RevenueReportItem, error) {
	log.Printf("⚙️ Service: Fetching revenue report from %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	var report []models.RevenueReportItem
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	query := `
		SELECT
			to_char(date_trunc('day', p.payment_date), 'YYYY-MM-DD') AS period,
			SUM(p.amount) AS amount
		FROM payments p
		WHERE p.payment_status = 'Paid'
		  AND p.payment_date >= $1
		  AND p.payment_date <= $2
		GROUP BY date_trunc('day', p.payment_date)
		ORDER BY period ASC;
	`
	err := config.DB.Select(&report, query, startDate, endDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.RevenueReportItem{}, nil
		}
		log.Printf("❌ Service: Error fetching revenue report: %v", err)
		return nil, fmt.Errorf("database error fetching revenue report: %w", err)
	}
	log.Printf("✅ Service: Fetched %d records for revenue report.", len(report))
	return report, nil
}

func GetPopularCarsReport(limit int) ([]models.PopularCarReportItem, error) {
	log.Printf("⚙️ Service: Fetching popular cars report (limit %d)", limit)
	var report []models.PopularCarReportItem
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT
			r.car_id,
			c.brand,
			c.model,
			COUNT(r.id) AS rental_count
		FROM rentals r
		JOIN cars c ON r.car_id = c.id
		GROUP BY r.car_id, c.brand, c.model
		ORDER BY rental_count DESC
		LIMIT $1;
	`
	err := config.DB.Select(&report, query, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.PopularCarReportItem{}, nil
		}
		log.Printf("❌ Service: Error fetching popular cars report: %v", err)
		return nil, fmt.Errorf("database error fetching popular cars report: %w", err)
	}
	log.Printf("✅ Service: Fetched %d records for popular cars report.", len(report))
	return report, nil
}

func GetBranchPerformanceReport() ([]models.BranchPerformanceReportItem, error) {
	log.Println("⚙️ Service: Fetching branch performance report")
	var report []models.BranchPerformanceReportItem
	query := `
		SELECT
			b.id AS branch_id,
			b.name AS branch_name,
			COUNT(DISTINCT r.id) AS total_rentals,
			COALESCE(SUM(CASE WHEN p.payment_status = 'Paid' THEN p.amount ELSE 0 END), 0) AS total_revenue
		FROM branches b
		LEFT JOIN cars c ON b.id = c.branch_id
		LEFT JOIN rentals r ON c.id = r.car_id
		LEFT JOIN payments p ON r.id = p.rental_id
		GROUP BY b.id, b.name
		ORDER BY b.name ASC;
	`
	err := config.DB.Select(&report, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.BranchPerformanceReportItem{}, nil
		}
		log.Printf("❌ Service: Error fetching branch performance report: %v", err)
		return nil, fmt.Errorf("database error fetching branch performance report: %w", err)
	}
	log.Printf("✅ Service: Fetched %d records for branch performance report.", len(report))
	return report, nil
}
