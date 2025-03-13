package models

import "time"

type Review struct {
	ID         int       `db:"id" json:"id"`
	CustomerID int       `db:"customer_id" json:"customer_id"`
	CarID      int       `db:"car_id" json:"car_id"`
	Rating     int       `db:"rating" json:"rating"`
	Comment    string    `db:"comment" json:"comment"`
	ReviewDate time.Time `db:"review_date" json:"review_date"`
}
