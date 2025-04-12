package models

import "time"

type Employee struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" binding:"required"`
	Email     string    `db:"email" json:"email" binding:"required,email"`
	Password  string    `db:"password" json:"password,omitempty" binding:"required,min=6"`
	Role      string    `db:"role" json:"role" binding:"required,oneof=admin manager"` // Restricted roles
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
