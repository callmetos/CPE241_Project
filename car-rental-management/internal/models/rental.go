package models

import "time"

type CarSummary struct {
	Brand string `db:"brand" json:"brand"`
	Model string `db:"model" json:"model"`
}

type Rental struct {
	ID              int        `db:"id" json:"id"`
	CustomerID      int        `db:"customer_id" json:"customer_id"`
	CarID           int        `db:"car_id" json:"car_id"`
	BookingDate     *time.Time `db:"booking_date" json:"booking_date"`
	PickupDatetime  time.Time  `db:"pickup_datetime" json:"pickup_datetime"`
	DropoffDatetime time.Time  `db:"dropoff_datetime" json:"dropoff_datetime"`
	PickupLocation  *string    `db:"pickup_location" json:"pickup_location"`

	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Car CarSummary `db:"car" json:"car"`
}

type InitiateRentalInput struct {
	CarID           int       `json:"car_id" binding:"required"`
	PickupDatetime  time.Time `json:"pickup_datetime" binding:"required"`
	DropoffDatetime time.Time `json:"dropoff_datetime" binding:"required,gtfield=PickupDatetime"`
	PickupLocation  *string   `json:"pickup_location"`
}

type UpdateRentalStatusInput struct {
	Status string `json:"status" binding:"required,oneof=Confirmed Active Returned Cancelled"`
}
