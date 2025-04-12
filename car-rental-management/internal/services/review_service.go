package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"database/sql"
	"errors"

	// "fmt" // Not used
	"log"
	"strings"
	// Needed for model timestamps
)

// CreateReview allows a customer to review a completed rental
func CreateReview(review models.Review) (models.Review, error) {
	log.Printf("Attempting to create review for rental %d by customer %d", review.RentalID, review.CustomerID)

	// --- Validation ---
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
	// --- End Validation ---

	query := `INSERT INTO reviews (customer_id, rental_id, rating, comment)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id, created_at, updated_at` // Return generated fields

	// Scan the returned values into the review struct
	err = config.DB.QueryRow(query, review.CustomerID, review.RentalID, review.Rating, review.Comment).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)

	if err != nil {
		log.Printf("❌ Error inserting review for rental %d: %v", review.RentalID, err)
		// Check for UNIQUE constraint violation on rental_id
		if strings.Contains(err.Error(), "reviews_rental_id_key") {
			return models.Review{}, errors.New("a review for this rental already exists")
		}
		return models.Review{}, errors.New("failed to submit review")
	}

	log.Printf("✅ Review created successfully with ID: %d for rental %d", review.ID, review.RentalID)
	return review, nil // Return the review with generated fields
}

// GetReviewsByCar retrieves all reviews for a specific car
func GetReviewsByCar(carID int) ([]models.Review, error) {
	log.Println("Fetching reviews for car ID:", carID)
	var reviews []models.Review
	// Query joins reviews with rentals to filter by car_id
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

// GetReviewByRental retrieves the review for a specific rental
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

// DeleteReview allows customer (own) or admin/manager to delete a review
func DeleteReview(reviewID int, actorID int, actorRole string) error {
	log.Printf("Attempting to delete review %d by actor %d (role: %s)", reviewID, actorID, actorRole)

	// 1. Get review details (customer_id) to check ownership
	var reviewOwnerID int
	err := config.DB.Get(&reviewOwnerID, "SELECT customer_id FROM reviews WHERE id=$1", reviewID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("review not found")
		}
		log.Printf("❌ Error fetching review %d owner: %v", reviewID, err)
		return errors.New("failed to get review details")
	}

	// 2. Check permission
	allowed := false
	if actorRole == "customer" && actorID == reviewOwnerID {
		allowed = true // Customer deleting own review
	} else if actorRole == "admin" || actorRole == "manager" {
		allowed = true // Staff deleting any review
	}

	if !allowed {
		log.Printf("❌ Permission denied: Actor %d (role %s) cannot delete review %d owned by %d", actorID, actorRole, reviewID, reviewOwnerID)
		return errors.New("permission denied to delete this review")
	}

	// 3. Delete the review
	result, err := config.DB.Exec("DELETE FROM reviews WHERE id=$1", reviewID)
	if err != nil {
		log.Printf("❌ Error deleting review %d: %v", reviewID, err)
		return errors.New("failed to delete review")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ Could not verify rows affected for review delete %d: %v", reviewID, err)
		// Proceeding as delete might have worked anyway
	}
	if rowsAffected == 0 {
		// Should not happen if fetch worked, but check anyway
		return errors.New("review not found for deletion (or already deleted)")
	}

	log.Printf("✅ Review %d deleted successfully by actor %d (role %s)", reviewID, actorID, actorRole)
	return nil
}
