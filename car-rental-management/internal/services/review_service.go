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
	// "time" // Not directly used here, but models.Review uses it
)

func CreateReview(review models.Review) (models.Review, error) {
	log.Printf("Attempting to create review for rental %d by customer %d", review.RentalID, review.CustomerID)
	var rentalStatus string
	var actualCustomerID int
	err := config.DB.QueryRow("SELECT status, customer_id FROM rentals WHERE id=$1", review.RentalID).Scan(&rentalStatus, &actualCustomerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Review{}, errors.New("rental not found")
		}
		log.Printf("❌ DB error checking rental %d for review: %v", review.RentalID, err)
		return models.Review{}, errors.New("database error checking rental")
	}
	if actualCustomerID != review.CustomerID {
		return models.Review{}, errors.New("permission denied: you can only review your own rentals")
	}
	if rentalStatus != "Returned" {
		return models.Review{}, errors.New("cannot review rental: status is not 'Returned'")
	}
	query := `INSERT INTO reviews (customer_id, rental_id, rating, comment)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id, created_at, updated_at`
	err = config.DB.QueryRow(query, review.CustomerID, review.RentalID, review.Rating, review.Comment).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		log.Printf("❌ Error inserting review for rental %d: %v", review.RentalID, err)
		if strings.Contains(err.Error(), "reviews_rental_id_key") {
			return models.Review{}, errors.New("a review for this rental already exists")
		}
		return models.Review{}, errors.New("failed to submit review")
	}
	log.Printf("✅ Review created successfully with ID: %d for rental %d", review.ID, review.RentalID)
	return review, nil
}

func GetReviewsByCar(carID int) ([]models.Review, error) {
	log.Println("Fetching reviews for car ID:", carID)
	var reviews []models.Review
	query := `SELECT r.id, r.customer_id, r.rental_id, r.rating, r.comment, r.created_at, r.updated_at
			  FROM reviews r
			  JOIN rentals rn ON r.rental_id = rn.id
			  WHERE rn.car_id = $1
			  ORDER BY r.created_at DESC`
	err := config.DB.Select(&reviews, query, carID)
	if err != nil {
		log.Printf("❌ Error fetching reviews for car %d: %v", carID, err)
		return nil, errors.New("failed to fetch reviews")
	}
	log.Printf("✅ Fetched %d reviews for car %d", len(reviews), carID)
	return reviews, nil
}

func GetReviewByRental(rentalID int) (models.Review, error) {
	log.Println("Fetching review for rental ID:", rentalID)
	var review models.Review
	query := `SELECT id, customer_id, rental_id, rating, comment, created_at, updated_at
              FROM reviews WHERE rental_id=$1`
	err := config.DB.Get(&review, query, rentalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Review{}, errors.New("review not found for this rental")
		}
		log.Printf("❌ Error fetching review for rental %d: %v", rentalID, err)
		return models.Review{}, errors.New("failed to fetch review")
	}
	log.Printf("✅ Review fetched successfully for rental %d", rentalID)
	return review, nil
}

func DeleteReview(reviewID int, actorID int, actorRole string) error {
	log.Printf("Attempting to delete review %d by actor %d (role: %s)", reviewID, actorID, actorRole)
	var reviewOwnerID int
	err := config.DB.Get(&reviewOwnerID, "SELECT customer_id FROM reviews WHERE id=$1", reviewID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("review not found")
		}
		log.Printf("❌ Error fetching review %d owner: %v", reviewID, err)
		return errors.New("failed to get review details")
	}
	allowed := false
	if actorRole == "customer" && actorID == reviewOwnerID {
		allowed = true
	} else if actorRole == "admin" || actorRole == "manager" {
		allowed = true
	}
	if !allowed {
		log.Printf("❌ Permission denied: Actor %d (role %s) cannot delete review %d owned by %d", actorID, actorRole, reviewID, reviewOwnerID)
		return errors.New("permission denied to delete this review")
	}
	result, err := config.DB.Exec("DELETE FROM reviews WHERE id=$1", reviewID)
	if err != nil {
		log.Printf("❌ Error deleting review %d: %v", reviewID, err)
		return errors.New("failed to delete review")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ Could not verify rows affected for review delete %d: %v", reviewID, err)
	}
	if rowsAffected == 0 {
		return errors.New("review not found for deletion (or already deleted)")
	}
	log.Printf("✅ Review %d deleted successfully by actor %d (role %s)", reviewID, actorID, actorRole)
	return nil
}

// GetAllReviewsPaginated fetches all reviews with details for admin panel
func GetAllReviewsPaginated(filters models.ReviewFiltersWithPagination) (models.PaginatedAdminReviewsResponse, error) {
	var response models.PaginatedAdminReviewsResponse
	response.Reviews = []models.AdminReviewView{} // Initialize

	queryBuilder := strings.Builder{}
	countQueryBuilder := strings.Builder{}
	args := []interface{}{}
	paramCount := 1

	baseSelect := `
		SELECT
			rev.id,
			rev.rental_id,
			rev.rating,
			rev.comment,
			rev.created_at AS review_created_at,
			rev.customer_id,
			cust.name AS customer_name,
			r.car_id,
			ca.brand AS car_brand,
			ca.model AS car_model
		FROM reviews rev
		JOIN rentals r ON rev.rental_id = r.id
		JOIN customers cust ON rev.customer_id = cust.id
		JOIN cars ca ON r.car_id = ca.id
	`
	queryBuilder.WriteString(baseSelect)
	countQueryBuilder.WriteString("SELECT COUNT(rev.id) FROM reviews rev JOIN rentals r ON rev.rental_id = r.id JOIN customers cust ON rev.customer_id = cust.id JOIN cars ca ON r.car_id = ca.id")

	var conditions []string
	if filters.Rating != nil && *filters.Rating >= 1 && *filters.Rating <= 5 {
		conditions = append(conditions, fmt.Sprintf("rev.rating = $%d", paramCount))
		args = append(args, *filters.Rating)
		paramCount++
	}
	if filters.CustomerID != nil && *filters.CustomerID > 0 {
		conditions = append(conditions, fmt.Sprintf("rev.customer_id = $%d", paramCount))
		args = append(args, *filters.CustomerID)
		paramCount++
	}
	if filters.CarID != nil && *filters.CarID > 0 {
		conditions = append(conditions, fmt.Sprintf("r.car_id = $%d", paramCount))
		args = append(args, *filters.CarID)
		paramCount++
	}
	if filters.Keyword != nil && *filters.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("rev.comment ILIKE $%d", paramCount))
		args = append(args, "%"+*filters.Keyword+"%")
		paramCount++
	}

	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		queryBuilder.WriteString(whereClause)
		countQueryBuilder.WriteString(whereClause)
	}

	// Count total records
	err := config.DB.QueryRow(countQueryBuilder.String(), args...).Scan(&response.TotalCount)
	if err != nil {
		log.Printf("Error counting all reviews: %v", err)
		return response, fmt.Errorf("failed to count reviews: %w", err)
	}

	// Sorting
	orderByClause := " ORDER BY rev.created_at DESC" // Default sort
	if filters.SortBy != "" {
		validSortByFields := map[string]string{
			"review_created_at": "rev.created_at",
			"rating":            "rev.rating",
			"customer_id":       "rev.customer_id",
			"car_id":            "r.car_id",
		}
		dbColumn, isValidField := validSortByFields[filters.SortBy]
		if isValidField {
			sortDir := "ASC"
			if strings.ToUpper(filters.SortDirection) == "DESC" {
				sortDir = "DESC"
			}
			orderByClause = fmt.Sprintf(" ORDER BY %s %s, rev.id %s", dbColumn, sortDir, sortDir)
		}
	}
	queryBuilder.WriteString(orderByClause)

	// Pagination
	if filters.Limit <= 0 {
		filters.Limit = 10 // Default limit for admin view
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}
	offset := (filters.Page - 1) * filters.Limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCount, paramCount+1))
	args = append(args, filters.Limit, offset)

	// Fetch reviews
	rows, err := config.DB.Queryx(queryBuilder.String(), args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, nil
		}
		log.Printf("Error fetching all paginated reviews: %v", err)
		return response, fmt.Errorf("failed to fetch reviews: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reviewView models.AdminReviewView
		if err := rows.StructScan(&reviewView); err != nil {
			log.Printf("Error scanning admin review view: %v", err)
			continue
		}
		response.Reviews = append(response.Reviews, reviewView)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating admin review rows: %v", err)
		return response, fmt.Errorf("error processing admin review rows: %w", err)
	}

	response.Page = filters.Page
	response.Limit = filters.Limit
	if response.TotalCount > 0 && response.Limit > 0 {
		response.TotalPages = int(math.Ceil(float64(response.TotalCount) / float64(response.Limit)))
	} else {
		response.TotalPages = 0
	}

	log.Printf("Service: Fetched %d admin reviews. Page: %d, Limit: %d, TotalItems: %d, TotalPages: %d", len(response.Reviews), response.Page, response.Limit, response.TotalCount, response.TotalPages)
	return response, nil
}
