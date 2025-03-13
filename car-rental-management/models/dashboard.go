package models

import "time"

type Dashboard struct {
	ID             int       `db:"id" json:"id"`
	EmployeeID     int       `db:"employee_id" json:"employee_id"`
	TotalRentals   int       `db:"total_rentals" json:"total_rentals"`
	TotalRevenue   float64   `db:"total_revenue" json:"total_revenue"`
	TotalCustomers int       `db:"total_customers" json:"total_customers"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
