package services

import (
	"car-rental-management/config"
	"car-rental-management/models"
	"log"
	"time"
)

// Create a new rental with pickup location
func CreateRental(rental models.Rental) error {
	log.Println("🔍 Creating a new rental:", rental)

	// Insert rental into the database
	_, err := config.DB.NamedExec(`
		INSERT INTO rentals (customer_id, car_id, rental_date, pickup_datetime, dropoff_datetime, pickup_location, status) 
		VALUES (:customer_id, :car_id, :rental_date, :pickup_datetime, :dropoff_datetime, :pickup_location, :status)`, rental)

	if err != nil {
		log.Println("❌ Error inserting rental:", err)
		return err
	}

	log.Println("✅ Rental created successfully!")
	return nil
}

// Get all rentals with execution time tracking
func GetRentals() ([]models.Rental, error) {
	var rentals []models.Rental
	start := time.Now()

	log.Println("🔍 Fetching rentals from database...")
	err := config.DB.Select(&rentals, "SELECT * FROM rentals")
	if err != nil {
		log.Println("❌ Error fetching rentals:", err)
		return nil, err
	}

	log.Printf("✅ Rentals fetched successfully in %v ms\n", time.Since(start).Milliseconds())
	return rentals, nil
}
