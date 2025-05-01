package models

import "time"

type Employee struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" `  // Removed binding for Create/Update
	Email     string    `db:"email" json:"email"` // Removed binding for Create/Update
	Password  string    `db:"password" json:"-"`  // Always omit password from JSON output
	Role      string    `db:"role" json:"role"`   // Removed binding for Create/Update
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreateEmployeeInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin manager"`
}

type UpdateEmployeeInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin manager"`
}
