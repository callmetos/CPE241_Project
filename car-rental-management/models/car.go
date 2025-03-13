package models

type Car struct {
	ID           int     `db:"id" json:"id"`
	Brand        string  `db:"brand" json:"brand"`
	Model        string  `db:"model" json:"model"`
	PricePerDay  float64 `db:"price_per_day" json:"price_per_day"`
	Availability bool    `db:"availability" json:"availability"`
	ParkingSpot  string  `db:"parking_spot" json:"parking_spot"` // ✅ New Field
}
