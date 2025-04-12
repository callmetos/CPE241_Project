package models

// DashboardData holds aggregated data for the dashboard
type DashboardData struct {
	TotalRentals   int     `db:"total_rentals" json:"total_rentals"`
	TotalRevenue   float64 `db:"total_revenue" json:"total_revenue"`
	TotalCustomers int     `db:"total_customers" json:"total_customers"`
	// Removed employee_id and updated_at as this is calculated dynamically
}
