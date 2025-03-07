// Rental struct
package models

type Rental struct {
	ID              int    `json:"id"`
	CustomerID      int    `json:"customer_id"`
	CarID           int    `json:"car_id"`
	PickupDatetime  string `json:"pickup_datetime"`
	DropoffDatetime string `json:"dropoff_datetime"`
	Status          string `json:"status"`
}
