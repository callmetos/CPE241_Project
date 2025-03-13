package services

import (
	"car-rental-management/config"
	"car-rental-management/models"
	"log"
)

// AddCar inserts a new car into the database
func AddCar(car models.Car) error { // ✅ Ensure this function is exported (capital A)
	log.Println("🔍 Adding a new car:", car)

	_, err := config.DB.NamedExec(`
		INSERT INTO cars (brand, model, price_per_day, availability, parking_spot) 
		VALUES (:brand, :model, :price_per_day, :availability, :parking_spot)`, car)

	if err != nil {
		log.Println("❌ Error inserting car:", err)
		return err
	}

	log.Println("✅ Car added successfully!")
	return nil
}

// GetAvailableCars retrieves all available cars from the database
func GetAvailableCars() ([]models.Car, error) {
	var cars []models.Car
	log.Println("🔍 Fetching available cars from the database...")

	err := config.DB.Select(&cars, "SELECT id, brand, model, price_per_day, availability, parking_spot FROM cars WHERE availability=true")
	if err != nil {
		log.Println("❌ Error fetching cars:", err)
		return nil, err
	}

	log.Println("✅ Cars fetched successfully!")
	return cars, nil
}
