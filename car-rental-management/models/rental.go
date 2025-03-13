package models

import "time"

type Rental struct {
	ID              int       `db:"id" json:"id"`
	CustomerID      int       `db:"customer_id" json:"customer_id"`
	CarID           int       `db:"car_id" json:"car_id"`
	RentalDate      string    `db:"rental_date" json:"rental_date"`
	PickupDatetime  time.Time `db:"pickup_datetime" json:"pickup_datetime"`
	DropoffDatetime time.Time `db:"dropoff_datetime" json:"dropoff_datetime"`
	PickupLocation  string    `db:"pickup_location" json:"pickup_location"`
	Status          string    `db:"status" json:"status"`
}
