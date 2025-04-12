package models

import "time"

type Car struct {
	ID           int       `db:"id" json:"id"`
	Brand        string    `db:"brand" json:"brand" binding:"required"`
	Model        string    `db:"model" json:"model" binding:"required"`
	PricePerDay  float64   `db:"price_per_day" json:"price_per_day" binding:"required,gt=0"`
	Availability bool      `db:"availability" json:"availability"`
	ParkingSpot  *string   `db:"parking_spot" json:"parking_spot"` // Pointer for nullable
	BranchID     int       `db:"branch_id" json:"branch_id" binding:"required"`
	ImageURL     *string   `db:"image_url" json:"image_url"` // Pointer for nullable
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
