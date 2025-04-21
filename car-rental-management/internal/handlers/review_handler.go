package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"database/sql" // Import sql package for ErrNoRows
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings" // Needed for error checking

	"github.com/gin-gonic/gin"
	"github.com/lib/pq" // For specific DB errors
)

// SubmitReview handles submitting a review for a specific rental (by customer)
func SubmitReview(c *gin.Context) {
	rentalIDStr := c.Param("id") // Use "id" consistently based on router fix
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	customerIDInterface, custExists := c.Get("customer_id") // Get customer ID from AuthMiddleware context
	// Check if customer_id exists AND is of type int
	customerID, ok := customerIDInterface.(int)
	if !custExists || !ok || customerID <= 0 { // Combined check
		log.Printf("🔥 Invalid customer_id (%v, exists: %v, ok: %v) in context for SubmitReview", customerIDInterface, custExists, ok)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required or invalid token data"})
		return
	}

	var input models.CreateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	review := models.Review{
		RentalID:   rentalID,
		CustomerID: customerID, // Set from authenticated user
		Rating:     input.Rating,
		Comment:    input.Comment,
	}

	createdReview, err := services.CreateReview(review)
	if err != nil {
		log.Printf("❌ Handler: Error submitting review for rental %d by customer %d: %v", rentalID, customerID, err)
		errMsg := err.Error()
		statusCode := http.StatusInternalServerError // Default to 500

		// This section uses errors.Is which is correct. The nilness warning is likely a false positive.
		if errors.Is(err, errors.New("rental not found")) {
			statusCode = http.StatusNotFound
			errMsg = "Rental not found" // Use clean message
		} else if errors.Is(err, errors.New("permission denied: you can only review your own rentals")) {
			statusCode = http.StatusForbidden
			errMsg = err.Error()
		} else if errors.Is(err, errors.New("cannot review rental: status is not 'Returned'")) {
			statusCode = http.StatusBadRequest // Bad request because prerequisite not met
			errMsg = err.Error()
		} else if errors.Is(err, errors.New("a review for this rental already exists")) || (err != nil && strings.Contains(errMsg, "reviews_rental_id_key")) { // Check for wrapped error too
			statusCode = http.StatusConflict // 409 Conflict
			errMsg = "A review for this rental already exists"
		} else if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { // Catch potential wrapped unique constraint error
			statusCode = http.StatusConflict
			errMsg = "A review for this rental already exists"
		} else if strings.Contains(errMsg, "invalid") { // Catch other potential validation errors
			statusCode = http.StatusBadRequest
		} else {
			errMsg = "Failed to submit review due to an internal error" // Generic fallback
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusCreated, createdReview)
}

// GetCarReviews handles fetching reviews for a specific car (public)
func GetCarReviews(c *gin.Context) {
	carIDStr := c.Param("id") // Use "id" consistently based on router fix
	carID, err := strconv.Atoi(carIDStr)
	if err != nil || carID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	reviews, err := services.GetReviewsByCar(carID)
	if err != nil {
		log.Printf("❌ Handler: Error fetching reviews for car %d: %v", carID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}
	// If reviews is empty, it's not an error, just return empty list
	c.JSON(http.StatusOK, reviews)
}

// GetRentalReview handles fetching the review for a specific rental
// Requires authentication (staff or customer who owns the rental).
func GetRentalReview(c *gin.Context) {
	rentalIDStr := c.Param("id") // Use "id" consistently
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	// --- Permission Check ---
	isAllowed := false
	actorRoleInterface, roleExists := c.Get("user_role") // role set by AuthMiddleware
	actorID := 0
	var rental models.Rental // To store fetched rental data

	if !roleExists {
		log.Println("❌ GetRentalReview: user_role not found in context.")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication information missing"})
		return
	}
	actorRole, _ := actorRoleInterface.(string) // Assume string if exists

	// Fetch rental details first to check ownership
	rental, err = services.GetRentalByID(rentalID) // Reuse existing service
	if err != nil {
		if errors.Is(err, errors.New("rental not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rental not found"})
		} else {
			log.Printf("❌ Error fetching rental %d for review lookup: %v", rentalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rental information"})
		}
		return
	}

	// Determine actor ID and check permission
	if actorRole == "admin" || actorRole == "manager" {
		isAllowed = true
		empIDInterface, empExists := c.Get("employee_id")
		if empExists {
			if empIDInt, ok := empIDInterface.(int); ok {
				actorID = empIDInt
			}
		}
	} else if actorRole == "customer" {
		custIDInterface, custExists := c.Get("customer_id") // Declaration of custExists
		if custExists {                                     // Usage of custExists
			if custIDInt, ok := custIDInterface.(int); ok {
				actorID = custIDInt
				if rental.CustomerID == actorID { // Check if customer owns the rental
					isAllowed = true
				}
			}
		}
	}

	if !isAllowed {
		log.Printf("⚠️ Permission Denied: Actor role '%s' (ID %d) tried to access review for rental %d owned by %d", actorRole, actorID, rentalID, rental.CustomerID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied to view this review"})
		return
	}
	// --- End Permission Check ---

	review, err := services.GetReviewByRental(rentalID)
	if err != nil {
		// Service returns "review not found for this rental" if no rows
		if errors.Is(err, errors.New("review not found for this rental")) || errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Review not found for this rental"})
		} else {
			log.Printf("❌ Handler: Error fetching review for rental %d: %v", rentalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review"})
		}
		return
	}
	c.JSON(http.StatusOK, review)
}

// DeleteReview handles deleting a review (customer own, or staff)
func DeleteReview(c *gin.Context) {
	reviewIDStr := c.Param("id") // Assuming route like /reviews/:id
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil || reviewID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	// Get user info from context to check permissions
	var actorID int
	var actorRole string
	var idSet bool

	// Prefer Employee check first
	employeeIDInterface, empExists := c.Get("employee_id")
	userRoleInterface, roleExists := c.Get("user_role") // Get role regardless

	if !roleExists {
		log.Println("❌ DeleteReview: user_role not found in context.")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication information missing"})
		return
	}
	roleStr, _ := userRoleInterface.(string) // Assume role exists if check passed

	if empExists {
		if empIDInt, ok := employeeIDInterface.(int); ok && empIDInt > 0 {
			actorID = empIDInt
			actorRole = roleStr // Should be "admin" or "manager"
			idSet = true
		}
	}

	// If not identified as employee, check if customer
	if !idSet {
		customerIDInterface, custExists := c.Get("customer_id") // <<<<<<<< Declaration of custExists
		if custExists {                                         // <<<<<<<< Usage of custExists
			if custIDInt, ok := customerIDInterface.(int); ok && custIDInt > 0 {
				actorID = custIDInt
				actorRole = roleStr // Should be "customer"
				idSet = true
			}
		}
	}

	if !idSet {
		// This happens if token is valid but contains neither valid employee_id nor customer_id
		custIDCheck, custCheckExists := c.Get("customer_id") // Re-check existence explicitly for logging
		log.Printf("❌ DeleteReview: Could not determine valid actor ID from context. Role: %s, empExists:%v, custExists:%v (value: %v)", roleStr, empExists, custCheckExists, custIDCheck)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token data"})
		return
	}

	err = services.DeleteReview(reviewID, actorID, actorRole)
	if err != nil {
		log.Printf("❌ Handler: Error deleting review %d by actor %d (%s): %v", reviewID, actorID, actorRole, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to delete review"

		errStr := err.Error()
		if errors.Is(err, errors.New("review not found")) {
			statusCode = http.StatusNotFound
			errMsg = errStr
		} else if errors.Is(err, errors.New("permission denied to delete this review")) {
			statusCode = http.StatusForbidden
			errMsg = errStr
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}
