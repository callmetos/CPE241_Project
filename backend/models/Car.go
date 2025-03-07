// Car struct
package models

type Car struct {
	ID           int     `json:"id"`
	Brand        string  `json:"brand"`
	Model        string  `json:"model"`
	PricePerDay  float64 `json:"price_per_day"`
	Availability bool    `json:"availability"`
}
