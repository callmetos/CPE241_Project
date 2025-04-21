package models

import "time"

type Rental struct {
	ID              int        `db:"id" json:"id"`
	CustomerID      int        `db:"customer_id" json:"customer_id"`           // Removed binding, set in service
	CarID           int        `db:"car_id" json:"car_id"`                     // Removed binding, set in service
	BookingDate     *time.Time `db:"booking_date" json:"booking_date"`         // Changed to *time.Time, DB default handles this
	PickupDatetime  time.Time  `db:"pickup_datetime" json:"pickup_datetime"`   // Removed binding
	DropoffDatetime time.Time  `db:"dropoff_datetime" json:"dropoff_datetime"` // Removed binding
	PickupLocation  *string    `db:"pickup_location" json:"pickup_location"`   // Pointer for nullable
	Status          string     `db:"status" json:"status"`                     // Removed binding, set in service
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// CreateRentalInput struct specifically for creating a rental by customer/staff
type CreateRentalInput struct {
	CustomerID      *int      `json:"customer_id"` // Staff must provide this if creating for customer, Customer ID taken from context
	CarID           int       `json:"car_id" binding:"required"`
	PickupDatetime  time.Time `json:"pickup_datetime" binding:"required"`
	DropoffDatetime time.Time `json:"dropoff_datetime" binding:"required,gtfield=PickupDatetime"` // Ensure dropoff is after pickup
	PickupLocation  *string   `json:"pickup_location"`                                            // Optional
}

// Input struct for updating rental status by staff
type UpdateRentalStatusInput struct {
	Status string `json:"status" binding:"required,oneof=Confirmed Active Returned Cancelled"`
}
