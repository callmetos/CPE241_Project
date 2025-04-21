package models

import "time"

// Customer struct represents the customer entity in the database.
type Customer struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`   // Removed binding, use Input structs
	Email     string    `db:"email" json:"email"` // Removed binding
	Phone     *string   `db:"phone" json:"phone"` // Pointer for nullable
	Password  string    `db:"password" json:"-"`  // Always omit from JSON output
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// RegisterCustomerInput struct for binding customer registration data.
type RegisterCustomerInput struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Phone    *string `json:"phone"`                             // Optional
	Password string  `json:"password" binding:"required,min=6"` // Password required for registration
}

// UpdateCustomerProfileInput struct for binding data when a customer updates their own profile.
type UpdateCustomerProfileInput struct {
	Name  string  `json:"name" binding:"required"`
	Phone *string `json:"phone"` // Optional
}

// UpdateCustomerByStaffInput struct for binding data when staff updates a customer.
// Staff might be allowed to update more fields, including email.
type UpdateCustomerByStaffInput struct {
	Name  string  `json:"name" binding:"required"`
	Email string  `json:"email" binding:"required,email"`
	Phone *string `json:"phone"` // Optional
	// Note: Staff typically cannot change password directly here.
}
