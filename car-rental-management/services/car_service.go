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

// UpdateCar updates car details
func UpdateCar(car models.Car) error {
	log.Println("🔄 Updating car:", car.ID)

	_, err := config.DB.NamedExec(`
		UPDATE cars SET brand=:brand, model=:model, price_per_day=:price_per_day, availability=:availability, parking_spot=:parking_spot
		WHERE id=:id`, car)

	if err != nil {
		log.Println("❌ Error updating car:", err)
		return err
	}

	log.Println("✅ Car updated successfully!")
	return nil
}

// DeleteCar removes a car from the database
func DeleteCar(carID int) error {
	log.Println("🗑 Deleting car with ID:", carID)

	_, err := config.DB.Exec("DELETE FROM cars WHERE id=$1", carID)
	if err != nil {
		log.Println("❌ Error deleting car:", err)
		return err
	}

	log.Println("✅ Car deleted successfully!")
	return nil
}
