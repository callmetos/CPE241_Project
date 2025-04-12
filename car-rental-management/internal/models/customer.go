package models

import "time"

type Customer struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" binding:"required"`
	Email     string    `db:"email" json:"email" binding:"required,email"`
	Phone     *string   `db:"phone" json:"phone"` // Pointer for nullable
	Password  string    `db:"password" json:"password,omitempty" binding:"omitempty,min=6"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Input struct for updating profile by customer
type UpdateCustomerProfileInput struct {
	Name  string  `json:"name" binding:"required"`
	Phone *string `json:"phone"`
}
