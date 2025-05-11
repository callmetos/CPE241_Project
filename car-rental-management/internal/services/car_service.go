package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
)

type CarFiltersWithPagination struct {
	Brand         *string
	Model         *string
	BranchID      *int
	MinPrice      *float64
	MaxPrice      *float64
	Availability  *bool
	Page          int
	Limit         int
	SortBy        string
	SortDirection string
}

type PaginatedCarsResponse struct {
	Cars       []models.Car `json:"cars"`
	TotalCount int          `json:"total_count"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	TotalPages int          `json:"total_pages"`
}

func GetCarsPaginated(filters CarFiltersWithPagination) (PaginatedCarsResponse, error) {
	var response PaginatedCarsResponse
	response.Cars = []models.Car{}

	queryBuilder := strings.Builder{}
	countQueryBuilder := strings.Builder{}
	args := []interface{}{}
	paramCount := 1

	queryBuilder.WriteString("SELECT c.id, c.brand, c.model, c.price_per_day, c.availability, c.parking_spot, c.branch_id, c.image_url, c.created_at, c.updated_at FROM cars c")
	countQueryBuilder.WriteString("SELECT COUNT(*) FROM cars c")

	var conditions []string

	if filters.Brand != nil && *filters.Brand != "" {
		conditions = append(conditions, fmt.Sprintf("c.brand ILIKE $%d", paramCount))
		args = append(args, "%"+*filters.Brand+"%")
		paramCount++
	}
	if filters.Model != nil && *filters.Model != "" {
		conditions = append(conditions, fmt.Sprintf("c.model ILIKE $%d", paramCount))
		args = append(args, "%"+*filters.Model+"%")
		paramCount++
	}
	if filters.BranchID != nil && *filters.BranchID > 0 {
		conditions = append(conditions, fmt.Sprintf("c.branch_id = $%d", paramCount))
		args = append(args, *filters.BranchID)
		paramCount++
	}
	if filters.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("c.price_per_day >= $%d", paramCount))
		args = append(args, *filters.MinPrice)
		paramCount++
	}
	if filters.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("c.price_per_day <= $%d", paramCount))
		args = append(args, *filters.MaxPrice)
		paramCount++
	}
	if filters.Availability != nil {
		conditions = append(conditions, fmt.Sprintf("c.availability = $%d", paramCount))
		args = append(args, *filters.Availability)
		paramCount++
	}

	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		queryBuilder.WriteString(whereClause)
		countQueryBuilder.WriteString(whereClause)
	}

	err := config.DB.QueryRow(countQueryBuilder.String(), args...).Scan(&response.TotalCount)
	if err != nil {
		log.Printf("Error counting cars: %v", err)
		return response, fmt.Errorf("failed to count cars: %w", err)
	}

	orderByClause := " ORDER BY c.id ASC"
	if filters.SortBy != "" {
		validSortByFields := map[string]bool{"id": true, "brand": true, "model": true, "price_per_day": true, "branch_id": true, "availability": true}
		if validSortByFields[filters.SortBy] {
			sortDir := "ASC"
			if strings.ToUpper(filters.SortDirection) == "DESC" {
				sortDir = "DESC"
			}
			orderByClause = fmt.Sprintf(" ORDER BY c.%s %s, c.id %s", filters.SortBy, sortDir, sortDir)
		}
	}
	queryBuilder.WriteString(orderByClause)

	if filters.Limit <= 0 {
		filters.Limit = 10
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}
	offset := (filters.Page - 1) * filters.Limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCount, paramCount+1))
	args = append(args, filters.Limit, offset)

	rows, err := config.DB.Queryx(queryBuilder.String(), args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, nil
		}
		log.Printf("Error fetching paginated cars: %v", err)
		return response, fmt.Errorf("failed to fetch cars: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var car models.Car
		if err := rows.StructScan(&car); err != nil {
			log.Printf("Error scanning car: %v", err)
			continue
		}
		response.Cars = append(response.Cars, car)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating car rows: %v", err)
		return response, fmt.Errorf("error processing car rows: %w", err)
	}

	response.Page = filters.Page
	response.Limit = filters.Limit
	if response.TotalCount > 0 && response.Limit > 0 {
		response.TotalPages = int(math.Ceil(float64(response.TotalCount) / float64(response.Limit)))
	} else {
		response.TotalPages = 0
	}
	if len(response.Cars) == 0 && response.Page > 1 && response.TotalPages > 0 && response.Page > response.TotalPages {
		log.Printf("Requested page %d is out of bounds (Total Pages: %d)", response.Page, response.TotalPages)
	}

	log.Printf("Service: Fetched %d cars. Page: %d, Limit: %d, TotalItems: %d, TotalPages: %d", len(response.Cars), response.Page, response.Limit, response.TotalCount, response.TotalPages)
	return response, nil
}

// GetCars is updated to use CarFiltersWithPagination or call GetCarsPaginated
// For this fix, we'll make it call GetCarsPaginated with a large limit
// to maintain its old behavior of fetching "all" cars if it's still used internally.
// The parameter type is changed to CarFiltersWithPagination to resolve the undefined: CarFilters error.
// If CarFilters was a distinct simple struct, you'd define it. But assuming it was for the old GetCars.
func GetCars(filters CarFiltersWithPagination) ([]models.Car, error) {
	// To mimic fetching all, set a very high limit.
	// Or, if this function is truly deprecated, it can be removed.
	// For now, let's assume it might be called internally and should try to fetch many.
	filters.Page = 1
	if filters.Limit == 0 { // If no limit was specified in this specific call
		filters.Limit = 10000 // A large number
	}
	if filters.SortBy == "" {
		filters.SortBy = "id"
	}
	if filters.SortDirection == "" {
		filters.SortDirection = "ASC"
	}

	result, err := GetCarsPaginated(filters)
	if err != nil {
		return nil, err
	}
	return result.Cars, nil
}

func AddCar(car models.Car) (models.Car, error) {
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
	var exists bool
	err := config.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM branches WHERE id=$1)", car.BranchID)
	if err != nil {
		log.Printf("Database error checking branch for car add: %v", err)
		return models.Car{}, errors.New("database error checking branch")
	}
	if !exists {
		return models.Car{}, fmt.Errorf("branch with ID %d does not exist", car.BranchID)
	}

	var insertedCar models.Car
	err = config.DB.QueryRowx(
		`INSERT INTO cars (brand, model, price_per_day, availability, parking_spot, branch_id, image_url)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		car.Brand, car.Model, car.PricePerDay, car.Availability, car.ParkingSpot, car.BranchID, car.ImageURL,
	).Scan(&insertedCar.ID, &insertedCar.CreatedAt, &insertedCar.UpdatedAt)

	if err != nil {
		log.Printf("Error inserting car into database: %v", err)
		return models.Car{}, errors.New("failed to add car to database")
	}
	car.ID = insertedCar.ID
	car.CreatedAt = insertedCar.CreatedAt
	car.UpdatedAt = insertedCar.UpdatedAt
	log.Printf("Car added successfully with ID: %d", car.ID)
	return car, nil
}

func GetCarByID(carID int) (models.Car, error) {
	var car models.Car
	if carID <= 0 {
		return models.Car{}, errors.New("invalid car ID")
	}
	query := `SELECT c.id, c.brand, c.model, c.price_per_day, c.availability, c.parking_spot, c.branch_id, c.image_url, c.created_at, c.updated_at
			  FROM cars c
			  WHERE c.id=$1`
	err := config.DB.Get(&car, query, carID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Car{}, errors.New("car not found")
		}
		log.Printf("Error fetching car by ID %d: %v", carID, err)
		return models.Car{}, errors.New("failed to fetch car")
	}
	return car, nil
}

func UpdateCar(car models.Car) (models.Car, error) {
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
	var branchExists bool
	err := config.DB.Get(&branchExists, "SELECT EXISTS(SELECT 1 FROM branches WHERE id=$1)", car.BranchID)
	if err != nil {
		log.Printf("Database error checking branch for car update: %v", err)
		return models.Car{}, errors.New("database error checking branch for update")
	}
	if !branchExists {
		return models.Car{}, fmt.Errorf("cannot update car, branch with ID %d does not exist", car.BranchID)
	}

	query := `
		UPDATE cars SET
			brand=:brand, model=:model, price_per_day=:price_per_day,
			availability=:availability, parking_spot=:parking_spot,
			branch_id=:branch_id, image_url=:image_url
		WHERE id=:id`
	result, err := config.DB.NamedExec(query, car)
	if err != nil {
		log.Printf("Error updating car ID %d: %v", car.ID, err)
		return models.Car{}, errors.New("failed to update car")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.Car{}, errors.New("car not found for update or no changes made")
	}
	updatedCar, fetchErr := GetCarByID(car.ID)
	if fetchErr != nil {
		log.Printf("Failed to fetch car after update ID %d: %v", car.ID, fetchErr)
		return car, nil
	}
	return updatedCar, nil
}

func DeleteCar(carID int) error {
	if carID <= 0 {
		return errors.New("invalid car ID for deletion")
	}
	result, err := config.DB.Exec("DELETE FROM cars WHERE id=$1", carID)
	if err != nil {
		if strings.Contains(err.Error(), "rentals_car_id_fkey") {
			return errors.New("cannot delete car: it has associated rentals")
		}
		log.Printf("Error deleting car ID %d: %v", carID, err)
		return errors.New("failed to delete car")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("car not found for deletion")
	}
	return nil
}
