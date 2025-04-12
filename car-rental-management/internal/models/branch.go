package models

import "time"

type Branch struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" binding:"required"`
	Address   *string   `db:"address" json:"address"` // Use pointer for nullable text
	Phone     *string   `db:"phone" json:"phone"`     // Use pointer for nullable varchar
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
