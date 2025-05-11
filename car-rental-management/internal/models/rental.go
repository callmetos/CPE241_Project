package models

import "time"

// CarSummary struct (ยังคงเดิม)
type CarSummary struct {
	Brand string `db:"brand" json:"brand"`
	Model string `db:"model" json:"model"`
}

// Rental struct (ยังคงเดิม)
type Rental struct {
	ID              int        `db:"id" json:"id"`
	CustomerID      int        `db:"customer_id" json:"customer_id"`
	CarID           int        `db:"car_id" json:"car_id"`
	BookingDate     *time.Time `db:"booking_date" json:"booking_date"` // Can be null if pending
	PickupDatetime  time.Time  `db:"pickup_datetime" json:"pickup_datetime"`
	DropoffDatetime time.Time  `db:"dropoff_datetime" json:"dropoff_datetime"`
	PickupLocation  *string    `db:"pickup_location" json:"pickup_location"`
	Status          string     `db:"status" json:"status"` // e.g., Pending, Booked, Confirmed, Active, Returned, Cancelled, Pending Verification
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
	Car             CarSummary `db:"car" json:"car"` // For embedding car brand and model
}

// InitiateRentalInput struct (ยังคงเดิม)
type InitiateRentalInput struct {
	CarID           int       `json:"car_id" binding:"required"`
	PickupDatetime  time.Time `json:"pickup_datetime" binding:"required"`
	DropoffDatetime time.Time `json:"dropoff_datetime" binding:"required,gtfield=PickupDatetime"`
	PickupLocation  *string   `json:"pickup_location"`
}

// UpdateRentalStatusInput struct (ยังคงเดิม)
type UpdateRentalStatusInput struct {
	Status string `json:"status" binding:"required,oneof=Confirmed Active Returned Cancelled"`
}

// --- Structs for Pagination and Filtering of Rentals ---

// RentalFiltersWithPagination struct สำหรับรับพารามิเตอร์การกรองและแบ่งหน้าสำหรับ Rentals
type RentalFiltersWithPagination struct {
	RentalID        *int       // กรองตาม rental_id
	CustomerID      *int       // กรองตาม customer_id
	CarID           *int       // กรองตาม car_id
	Status          *string    // กรองตาม status
	PickupDateAfter *time.Time // กรองตามวันที่รับรถ (หลังจากวันที่ระบุ)
	Page            int
	Limit           int
	SortBy          string // เช่น "id", "pickup_datetime", "status"
	SortDirection   string // "ASC" หรือ "DESC"
}

// PaginatedRentalsResponse struct สำหรับ response ที่มีการแบ่งหน้าของ Rental
type PaginatedRentalsResponse struct {
	Rentals    []Rental `json:"rentals"` // Array of Rental structs
	TotalCount int      `json:"total_count"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}
