// Rental struct
package models

type Rental struct {
	ID         int    `json:"id"`
	CustomerID int    `json:"customer_id"`
	CarID      int    `json:"car_id"`
	RentalDate string `json:"rental_date"`
	ReturnDate string `json:"return_date"`
	Status     string `json:"status"`
}
