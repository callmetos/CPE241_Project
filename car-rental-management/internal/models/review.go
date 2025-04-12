package models

import "time"

type Review struct {
	ID         int       `db:"id" json:"id"`
	CustomerID int       `db:"customer_id" json:"customer_id"`
	RentalID   int       `db:"rental_id" json:"rental_id"`
	Rating     int       `db:"rating" json:"rating" binding:"required,min=1,max=5"`
	Comment    *string   `db:"comment" json:"comment"` // Pointer for nullable
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// Input struct for creating a review
type CreateReviewInput struct {
	Rating  int     `json:"rating" binding:"required,min=1,max=5"`
	Comment *string `json:"comment"`
}
