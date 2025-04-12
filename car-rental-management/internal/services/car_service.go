package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	// Assuming time.Time in model for CreatedAt/UpdatedAt
)

// AddCar inserts a new car into the database after validation
func AddCar(car models.Car) (models.Car, error) {
	log.Println("🔍 Validating car data before adding:", car.Brand, car.Model)

	// --- Input Validation ---
	if strings.TrimSpace(car.Brand) == "" {
		return models.Car{}, errors.New("car brand cannot be empty")
	}
	if strings.TrimSpace(car.Model) == "" {
		return models.Car{}, errors.New("car model cannot be empty")
	}
	if car.PricePerDay <= 0 {
		return models.Car{}, errors.New("price per day must be greater than zero")
	}
	if car.BranchID <= 0 {
		return models.Car{}, errors.New("invalid branch ID")
	}
	// --- End Input Validation ---

	// Optional: Check if BranchID actually exists
	var exists bool
	err := config.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM branches WHERE id=$1)", car.BranchID)
	if err != nil {
		log.Printf("❌ Error checking branch existence for ID %d: %v", car.BranchID, err)
		return models.Car{}, errors.New("database error checking branch")
	}
	if !exists {
		return models.Car{}, fmt.Errorf("branch with ID %d does not exist", car.BranchID)
	}

	log.Println("✅ Car data validation passed. Adding car to database...")

	var insertedCar models.Car // Use a temporary struct to scan generated values
	// Use QueryRowx to get back generated fields reliably
	err = config.DB.QueryRowx(
		`INSERT INTO cars (brand, model, price_per_day, availability, parking_spot, branch_id, image_url)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`, // Added created_at, updated_at
		car.Brand, car.Model, car.PricePerDay, car.Availability, car.ParkingSpot, car.BranchID, car.ImageURL,
	).Scan(&insertedCar.ID, &insertedCar.CreatedAt, &insertedCar.UpdatedAt)

	if err != nil {
		log.Println("❌ Error inserting car:", err)
		// Consider checking for unique constraints if you add any (e.g., unique license plate)
		return models.Car{}, errors.New("failed to add car to database")
	}

	// Populate the rest of the returned struct based on input and returned values
	car.ID = insertedCar.ID
	car.CreatedAt = insertedCar.CreatedAt // Assign returned timestamp
	car.UpdatedAt = insertedCar.UpdatedAt // Assign returned timestamp

	log.Printf("✅ Car added successfully with ID: %d!", car.ID)
	return car, nil // Return the full car object
}

// CarFilters defines available filters for GetCars
type CarFilters struct {
	Brand        *string
	Model        *string
	BranchID     *int
	MinPrice     *float64
	MaxPrice     *float64
	Availability *bool
}

// GetCars retrieves cars, potentially filtered
func GetCars(filters CarFilters) ([]models.Car, error) {
	var cars []models.Car
	log.Println("🔍 Fetching cars with filters:", filters)

	baseQuery := `SELECT c.id, c.brand, c.model, c.price_per_day, c.availability, c.parking_spot, c.branch_id, c.image_url, c.created_at, c.updated_at
				  FROM cars c` // Add JOIN with branches if needed: JOIN branches b ON c.branch_id = b.id

	conditions := []string{}
	args := []interface{}{}
	argID := 1

	if filters.Brand != nil && *filters.Brand != "" {
		conditions = append(conditions, fmt.Sprintf("c.brand ILIKE $%d", argID)) // Case-insensitive search
		args = append(args, "%"+*filters.Brand+"%")
		argID++
	}
	if filters.Model != nil && *filters.Model != "" {
		conditions = append(conditions, fmt.Sprintf("c.model ILIKE $%d", argID))
		args = append(args, "%"+*filters.Model+"%")
		argID++
	}
	if filters.BranchID != nil {
		conditions = append(conditions, fmt.Sprintf("c.branch_id = $%d", argID))
		args = append(args, *filters.BranchID)
		argID++
	}
	if filters.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("c.price_per_day >= $%d", argID))
		args = append(args, *filters.MinPrice)
		argID++
	}
	if filters.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("c.price_per_day <= $%d", argID))
		args = append(args, *filters.MaxPrice)
		argID++
	}
	if filters.Availability != nil {
		conditions = append(conditions, fmt.Sprintf("c.availability = $%d", argID))
		args = append(args, *filters.Availability)
		argID++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY c.brand, c.model" // Default ordering

	// Use sqlx.Select for dynamic query arguments
	err := config.DB.Select(&cars, query, args...)
	if err != nil {
		log.Println("❌ Error fetching filtered cars:", err)
		return nil, errors.New("failed to fetch cars")
	}

	log.Printf("✅ Fetched %d cars successfully!", len(cars))
	return cars, nil
}

// GetCarByID retrieves a single car by ID, possibly with branch info
func GetCarByID(carID int) (models.Car, error) {
	var car models.Car
	log.Println("🔍 Fetching car by ID:", carID)
	query := `SELECT c.id, c.brand, c.model, c.price_per_day, c.availability, c.parking_spot, c.branch_id, c.image_url, c.created_at, c.updated_at
			  FROM cars c
			  WHERE c.id=$1` // Add JOIN with branches if needed to get branch name etc.
	err := config.DB.Get(&car, query, carID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ Car with ID %d not found.", carID)
			return models.Car{}, errors.New("car not found")
		}
		log.Printf("❌ Error fetching car %d: %v", carID, err)
		return models.Car{}, errors.New("failed to fetch car")
	}
	log.Printf("✅ Car %d fetched successfully!", carID)
	return car, nil
}

// UpdateCar updates car details
func UpdateCar(car models.Car) (models.Car, error) {
	log.Println("🔄 Validating car data before updating car:", car.ID)
	// Validation
	if car.ID <= 0 {
		return models.Car{}, errors.New("invalid car ID for update")
	}
	if strings.TrimSpace(car.Brand) == "" {
		return models.Car{}, errors.New("car brand cannot be empty")
	}
	if strings.TrimSpace(car.Model) == "" {
		return models.Car{}, errors.New("car model cannot be empty")
	}
	if car.PricePerDay <= 0 {
		return models.Car{}, errors.New("price per day must be greater than zero")
	}
	if car.BranchID <= 0 {
		return models.Car{}, errors.New("invalid branch ID")
	}
	// Optional: Check if BranchID actually exists (like in AddCar)
	// ...

	log.Println("✅ Car data validation passed. Updating car in database...")

	// Use NamedExec for updating based on struct fields
	query := `
		UPDATE cars SET
			brand=:brand, model=:model, price_per_day=:price_per_day,
			availability=:availability, parking_spot=:parking_spot,
			branch_id=:branch_id, image_url=:image_url
			-- updated_at is handled by trigger
		WHERE id=:id`
	result, err := config.DB.NamedExec(query, car)
	if err != nil {
		log.Println("❌ Error updating car:", err)
		return models.Car{}, errors.New("failed to update car")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.Car{}, errors.New("car not found for update")
	}

	log.Println("✅ Car updated successfully!")
	// Fetch the updated car data to return it with the new updated_at
	updatedCar, fetchErr := GetCarByID(car.ID)
	if fetchErr != nil {
		log.Printf("⚠️ Failed to fetch updated car data after update for ID %d: %v", car.ID, fetchErr)
		return car, nil // Return the input data as fallback
	}
	return updatedCar, nil
}

// DeleteCar removes a car from the database
func DeleteCar(carID int) error {
	log.Println("🗑 Deleting car with ID:", carID)
	// Optional: Check if car has active rentals before deleting? DB FK constraint handles this if set to RESTRICT.
	result, err := config.DB.Exec("DELETE FROM cars WHERE id=$1", carID)
	if err != nil {
		log.Printf("❌ Error deleting car %d: %v", carID, err)
		// Check for FK violation if RESTRICT is used
		if strings.Contains(err.Error(), "violates foreign key constraint") && strings.Contains(err.Error(), "rentals_car_id_fkey") {
			return errors.New("cannot delete car: it has associated rentals")
		}
		return errors.New("failed to delete car")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("car not found for deletion")
	}
	log.Println("✅ Car deleted successfully!")
	return nil
}
