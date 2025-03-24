package models

// DashboardData struct for statistics
type DashboardData struct {
	TotalRentals   int     `json:"total_rentals" db:"total_rentals"`
	TotalRevenue   float64 `json:"total_revenue" db:"total_revenue"`
	TotalCustomers int     `json:"total_customers" db:"total_customers"`
}
