package models

// DashboardData struct สำหรับ Admin Dashboard
type DashboardData struct {
	TotalRentals       int     `db:"total_rentals" json:"total_rentals"`
	TotalRevenue       float64 `db:"total_revenue" json:"total_revenue"`
	TotalCustomers     int     `db:"total_customers" json:"total_customers"`
	TotalCars          int     `db:"total_cars" json:"total_cars"`                     // เพิ่ม: จำนวนรถทั้งหมด
	TotalAvailableCars int     `db:"total_available_cars" json:"total_available_cars"` // มีอยู่แล้ว
	UnavailableCars    int     `db:"unavailable_cars" json:"unavailable_cars"`         // เพิ่ม: จำนวนรถที่ไม่ว่าง
	TotalBranches      int     `db:"total_branches" json:"total_branches"`             // มีอยู่แล้ว
}

// PublicStatsData struct สำหรับข้อมูลสถิติสาธารณะ (ยังคงเดิม)
type PublicStatsData struct {
	TotalAvailableCars int `db:"total_available_cars" json:"total_available_cars"`
	TotalBranches      int `db:"total_branches" json:"total_branches"`
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
