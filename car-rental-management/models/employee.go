package models

// Employee struct defines the fields for an employee
type Employee struct {
	ID       int    `db:"id" json:"id"`
	Name     string `db:"name" json:"name" binding:"required"`
	Email    string `db:"email" json:"email" binding:"required,email"`
	Password string `db:"password" json:"password,omitempty" binding:"required,min=6"`
	Role     string `db:"role" json:"role"`
}
