package models

type DashboardData struct {
	TotalRentals   int     `db:"total_rentals" json:"total_rentals"`
	TotalRevenue   float64 `db:"total_revenue" json:"total_revenue"`
	TotalCustomers int     `db:"total_customers" json:"total_customers"`
}

type RevenueReportItem struct {
	Period string  `db:"period" json:"period"`
	Amount float64 `db:"amount" json:"amount"`
}

type PopularCarReportItem struct {
	CarID       int    `db:"car_id" json:"car_id"`
	Brand       string `db:"brand" json:"brand"`
	Model       string `db:"model" json:"model"`
	RentalCount int    `db:"rental_count" json:"rental_count"`
}

type BranchPerformanceReportItem struct {
	BranchID     int     `db:"branch_id" json:"branch_id"`
	BranchName   string  `db:"branch_name" json:"branch_name"`
	TotalRentals int     `db:"total_rentals" json:"total_rentals"`
	TotalRevenue float64 `db:"total_revenue" json:"total_revenue"`
}
