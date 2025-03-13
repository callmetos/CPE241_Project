package models

import "time"

type Payment struct {
	ID            int       `db:"id" json:"id"`
	RentalID      int       `db:"rental_id" json:"rental_id"`
	Amount        float64   `db:"amount" json:"amount"`
	PaymentDate   time.Time `db:"payment_date" json:"payment_date"`
	PaymentStatus string    `db:"payment_status" json:"payment_status"`
}
