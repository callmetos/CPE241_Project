package models

import "time"

type Rental struct {
	ID              int       `db:"id" json:"id"`
	CustomerID      int       `db:"customer_id" json:"customer_id" binding:"required"`
	CarID           int       `db:"car_id" json:"car_id" binding:"required"`
	BookingDate     *string   `db:"booking_date" json:"booking_date"` // Use pointer for Date default
	PickupDatetime  time.Time `db:"pickup_datetime" json:"pickup_datetime" binding:"required"`
	DropoffDatetime time.Time `db:"dropoff_datetime" json:"dropoff_datetime" binding:"required,gtfield=PickupDatetime"` // Ensure dropoff is after pickup
	PickupLocation  *string   `db:"pickup_location" json:"pickup_location"`                                             // Pointer for nullable
	Status          string    `db:"status" json:"status" binding:"required,oneof=Booked Confirmed Active Returned Cancelled"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// Input struct specifically for creating a rental by customer/staff
type CreateRentalInput struct {
	CustomerID      *int      `json:"customer_id"` // Staff must provide this if creating for customer
	CarID           int       `json:"car_id" binding:"required"`
	PickupDatetime  time.Time `json:"pickup_datetime" binding:"required"`
	DropoffDatetime time.Time `json:"dropoff_datetime" binding:"required,gtfield=PickupDatetime"`
	PickupLocation  *string   `json:"pickup_location"`
}

// Input struct for updating rental status by staff
type UpdateRentalStatusInput struct {
	Status string `json:"status" binding:"required,oneof=Confirmed Active Returned Cancelled"`
}
